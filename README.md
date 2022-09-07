# maven-index-reader-go


## What?
A basic port of [this utility](https://github.com/apache/maven-indexer/tree/master/indexer-reader) to Go. Includes support for full or incremental updates starting after a supplied last-successfully-consumed chunk ID or RFC3339 chunk timestamp, filtering for various record types (only `ARTIFACT_ADD` and `ARTIFACT_REMOVE` are typically useful) and output in JSON or CSV formats to a local file or `stdout`.

There is an example binary [here](https://github.com/elireisman/maven-index-reader-go/blob/main/cmd/main.go) for dumping the Maven Central index that you can build by running `make` from the checkout root. Following that example, you can use the [public packages](https://github.com/elireisman/maven-index-reader-go/tree/main/pkg) as utility libraries to compose your own tool to parse other remote or local indices.

This isn't production quality yet. It's a PoC, and could benefit from some refactoring, improvement of various hackery, and better test coverage. I might get to _some_ of that in the near future. That said, I've been using it to scan the full and incremental chunks of the Maven Central index, and a few small test indices, without incident.


#### Usage Example
```bash
# Dump Maven Central add/remove records from all
# incremental index updates published _after_ chunk 768.
$ make
$ bin/index_reader --after 768 --mode after-chunk --format json > index.dump 

# Example output
$ head -10 index.dump
[
{"artifactId":"matlib","description":"Version-independent Material library for Bukkit 1.8+","fileExtension":"jar","fileModified":"2022-09-03T00:10:10Z","fileSize":196811,"groupId":"xyz.wasabicodes","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"MaterialLib","packaging":"jar","recordModified":"2022-09-04T07:39:11.839Z","recordType":"artifact_add","sha1":"38bb5a445e9aa5a38581743ede58f46c0f1ce321","version":"1.1.2"},
{"artifactId":"matlib","description":"Version-independent Material library for Bukkit 1.8+","fileExtension":"jar","fileModified":"2022-09-02T18:15:49Z","fileSize":196845,"groupId":"xyz.wasabicodes","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"MaterialLib","packaging":"jar","recordModified":"2022-09-04T07:39:11.94Z","recordType":"artifact_add","sha1":"6b725000db7afc05023d4b689a2e21cd3f3307a0","version":"1.1.1"},
{"artifactId":"matlib","description":"Version-independent Material library for Bukkit 1.8+","fileExtension":"jar","fileModified":"2022-09-02T17:30:32Z","fileSize":196485,"groupId":"xyz.wasabicodes","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"MaterialLib","packaging":"jar","recordModified":"2022-09-04T07:39:12.023Z","recordType":"artifact_add","sha1":"28d8997f5ac1f503d2d9620ffd385c0e5e212d62","version":"1.1.0"},
{"artifactId":"ens-client","description":"A simple and naive java wrapper library for retrieving \nENS-related information from an Ethereum node. This library is built on\ntop of the excellent web3j project (https://github.com/web3j/web3j).","fileExtension":"pom.sha512","fileModified":"2022-09-04T00:15:16Z","fileSize":3489,"groupId":"xyz.seleya.product","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"ens-client","packaging":"pom","recordModified":"2022-09-04T07:39:22.081Z","recordType":"artifact_add","sha1":"9ac67971d82fa2498ae6af66c1cc27bfea8533e0","version":"0.0.2"},
{"artifactId":"ens-client","description":"A simple and naive java wrapper library for retrieving \nENS-related information from an Ethereum node. This library is built on\ntop of the excellent web3j project (https://github.com/web3j/web3j).","fileExtension":"pom.sha512","fileModified":"2022-09-03T15:03:35Z","fileSize":3489,"groupId":"xyz.seleya.product","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"ens-client","packaging":"pom","recordModified":"2022-09-04T07:39:22.202Z","recordType":"artifact_add","sha1":"697d6b2ef68b95dcf6704a8fa6033e8dc071a86c","version":"0.0.1"},
{"artifactId":"trivial-chunk","description":"A library of trivial codes.","fileExtension":"pom.sha512","fileModified":"2022-08-29T05:44:31Z","fileSize":128,"groupId":"xyz.ronella.casual","hasJavadoc":true,"hasSignature":false,"hasSources":true,"name":"Trivial Chunk","packaging":"pom.sha512","recordModified":"2022-09-04T07:39:23.614Z","recordType":"artifact_add","version":"2.14.0"},
{"artifactId":"siths","description":"Coroutines-based Redis client library","fileExtension":"pom.sha512","fileModified":"2022-09-02T14:46:12Z","fileSize":128,"groupId":"xyz.haff","hasJavadoc":true,"hasSignature":false,"hasSources":true,"name":"siths","packaging":"pom.sha512","recordModified":"2022-09-04T07:39:48.415Z","recordType":"artifact_add","version":"0.9.0"},
{"artifactId":"koy","description":"Random assortment of Kotlin utilities and extensions.","fileExtension":"pom.sha512","fileModified":"2022-08-27T11:53:05Z","fileSize":128,"groupId":"xyz.haff","hasJavadoc":true,"hasSignature":false,"hasSources":true,"name":"koy","packaging":"pom.sha512","recordModified":"2022-09-04T07:39:48.759Z","recordType":"artifact_add","version":"0.5.1"},
{"artifactId":"koy","description":"Random assortment of Kotlin utilities and extensions.","fileExtension":"pom.sha512","fileModified":"2022-08-27T10:25:52Z","fileSize":128,"groupId":"xyz.haff","hasJavadoc":true,"hasSignature":false,"hasSources":true,"name":"koy","packaging":"pom.sha512","recordModified":"2022-09-04T07:39:48.835Z","recordType":"artifact_add","version":"0.5.0"},
```

## Why?
I know, I know...don't worry, I have my reasons :)
