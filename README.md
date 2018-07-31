# CVDemo

Coordinate Values Demo

## Approach

The file is read line by line and parsed into a Trie like structure consisting of nested hashtables.

This approach was selected to conserve space since the trie has the potential to grow exponentially with each
level.

Each trie node has a list of the CIDs (unique city IDs) that have the prefix denoted by the trie to that point.

## Note

The text file has an unknown encoding for extended characters (eg. Trois Rivières == Trois Rivi√®res) likely due to bad source file.

## Example

```text
GET /suggestions?q=Londo&latitude=43.70011&longitude=-79.4163

{
  "suggestions": [
    {
      "name": "London, ON, Canada",
      "latitude": "42.98339",
      "longitude": "-81.23304",
      "score": 0.9
    },
    {
      "name": "London, OH, USA",
      "latitude": "39.88645",
      "longitude": "-83.44825",
      "score": 0.5
    },
    {
      "name": "London, KY, USA",
      "latitude": "37.12898",
      "longitude": "-84.08326",
      "score": 0.5
    },
    {
      "name": "Londontowne, MD, USA",
      "latitude": "38.93345",
      "longitude": "-76.54941",
      "score": 0.3
    }
  ]
}
```