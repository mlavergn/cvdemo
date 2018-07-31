import io
import json
from http.server import BaseHTTPRequestHandler, HTTPServer
from urlparse import urlparse
import unicodedata

F_NAME = 'name'
F_LONGITUDE = 'longitude'
F_LATITUDE = 'latitude'
F_COUNTRY = 'country'
F_REGION = 'region'
F_SCORE = 'score'

class CVTrie:

    class CVTrieNode:
        def __init__(self, char, cid):
            self.char = char
            self.ids = set()
            self.ids.add(cid)
            self.nodes = {}

        def __repr__(self):
            return '%s %s %s' % (self.char, self.ids, self.nodes)

    def __init__(self):
        print('CVTrie::init')
        self.cities = {}
        self.nodes = {}

    def __repr__(self):
        return '%s' % (self.nodes.keys())

    def find(self, prefix, latitude, longitude):
        print('CVTrie::find')
        result = []

        # walk the trie until exhaustion of the prefix
        node = self
        for c in prefix.lower():
            node = node.nodes[c]

        # process the CIDs
        for cid in node.ids:
            city = self.cities[cid]
            resultNode = {
                F_NAME: '%s, %s, %s' % (city[F_NAME], city[F_REGION], city[F_COUNTRY]),
                F_LATITUDE: str(city[F_LATITUDE]),
                F_LONGITUDE: str(city[F_LONGITUDE])
            }
            resultNode[F_SCORE] = self.score(cid, prefix, latitude, longitude)
            result.append(resultNode)

        # sort in reverse order as per spec
        result.sort(cmp=lambda a, b: 1 if a[F_SCORE] < b[F_SCORE] else -1)

        return result

    def score(self, cid, query, latitude, longitude):
        print('CVTrie::score')
        city = self.cities[cid]
        print('CVTrie::score parameters', city, latitude, longitude)

        # name 40%
        divisor = max(len(query), len(city)) / min(len(query), len(city))
        prefField = min(40.0 / divisor, 40.0)

        # latitude 30%
        divisor = abs(latitude - city[F_LATITUDE])
        latField = min(30.0 / divisor, 30.0)

        # longitude 30%
        divisor = abs(longitude - city[F_LONGITUDE])
        longField = min(30.0 / divisor, 30.0)

        print('CVTrie::score components', prefField, latField, longField)
        return round((prefField + latField + longField) / 100.0, 1)

    def add(self, cid, name, asciiName, latitude, longitude, country, region):
        self.cities[cid] = {
            F_NAME: name,
            F_LONGITUDE: longitude,
            F_LATITUDE: latitude,
            F_COUNTRY: country,
            F_REGION: region
        }

        node = self
        indexable = list('abcdefghijklmnopqrstuvwxyz')
        extended = list('-\'., ()1')
        for c in asciiName.lower():
            if c not in indexable:
                if c not in extended:
                    print('Unknown character [%s] in [%s]' % (c, name))
                continue
            child = node.nodes.get(c, None)
            if child is None:
                child = self.CVTrieNode(c, cid)
                node.nodes[c] = child
            else:
                child.ids.add(cid)
            node = child

class CVLoader:
    def __init__(self, trie):
        print('CVLoader::init')

        regionLookup = {
            '01': 'AB',
            '02': 'BC',
            '03': 'MB',
            '04': 'NB',
            '05': 'NL',
            '07': 'NS',
            '08': 'ON',
            '09': 'PE',
            '10': 'QC',
            '11': 'SK',
            '12': 'YT',
            '13': 'NT',
            '14': 'NU'
        }

        try:
            self.legend = None
            with io.open('cities_canada-usa.tsv', mode='r', encoding='utf-8') as lines:
                for line in lines:
                    fields = line.strip().split('\t')
                    if self.legend is None:
                        # not used, for future consideration
                        self.legend = fields
                    else:
                        # fields: 0: id | 1: name | 2: ascii | 3: altname | 4: lat | 5: long | 8: country | 10: region
                        if fields[8] == 'US':
                            region = fields[10]
                        else:
                            region = regionLookup.get(fields[10], '')
                            if not region:
                                print('Failed region code lookup:', fields)
                                continue
                        if fields[8] == 'US':
                            country = 'USA'
                        else:
                            country = 'Canada'
                        trie.add(fields[0], fields[1], fields[2], float(fields[4]), float(fields[5]), country, region)
        except ValueError:
            print('Could not load data')

class CVRequestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        print('CVRequestHandler::do_GET url:[%s]' % self.path)
        query = urlparse(self.path).query
        params = {}
        if query:
            paramKVs = query.split('&')
            if paramKVs:
                params = dict(paramKV.split('=') for paramKV in paramKVs)

        # convert non-ascii characters
        prefix = params.get('q', None)
        prefix = bytes(prefix).decode('utf-8', 'strict')
        prefix = unicodedata.normalize('NFD', prefix).encode('ascii', 'ignore').decode('utf-8')

        # if not provided, use a minimal latitute or longitude
        latitude = float(params.get(F_LATITUDE, '0.00001'))
        longitude = float(params.get(F_LONGITUDE, '0.00001'))

        results = vtrie.find(prefix, latitude, longitude)

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        response = {
            'suggestions': results
        }
        self.wfile.write(json.dumps(response).encode('utf-8'))
        self.wfile.close()

class CVService:
    def start(self, handlerClass=CVRequestHandler):
        print('CVSerice::start')
        server_address = ('', 8080)
        httpd = HTTPServer(server_address, handlerClass)
        saddr = httpd.socket.getsockname()
        print('Serving Endpoint at [%s:%i]' % (saddr[0], saddr[1]))
        httpd.serve_forever()

if __name__ == '__main__':
    vtrie = CVTrie()
    CVLoader(vtrie)
    svc = CVService()
    svc.start()