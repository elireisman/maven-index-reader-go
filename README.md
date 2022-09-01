# maven-index-reader-go


## What?
A basic port of [this utility](https://github.com/apache/maven-indexer/tree/master/indexer-reader) to Go. Includes support for full or incremental updates starting after a supplied last-successfully-consumed chunk ID or RFC3339 chunk timestamp, filtering for various record types (only `ARTIFACT_ADD` and `ARTIFACT_REMOVE` are typically useful) and output in JSON or CSV formats to a local file or `stdout`.

There is an example binary [here](https://github.com/elireisman/maven-index-reader-go/blob/main/cmd/main.go) for dumping the Maven Central index that you can build by running `make` from the checkout root. Following that example, you can use the [public packages](https://github.com/elireisman/maven-index-reader-go/tree/main/pkg) as utility libraries to compose your own dumper for other remote or local indices.

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
{"artifactId":"spark-data-standardization_2.12","description":"95e60d6dcbd422b678029c8724c04b37cea38519","fileExtension":"jar","fileModified":"2022-08-18T16:57:58Z","fileSize":500330,"groupId":"za.co.absa","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.1"},
{"artifactId":"spark-data-standardization_2.12","description":"8347df993825619463c27b15a8591a77938e8739","fileExtension":"jar","fileModified":"2022-08-11T09:18:35Z","fileSize":500386,"groupId":"za.co.absa","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.0"},
{"artifactId":"spark-data-standardization_2.11","description":"472fff2e9b3c8e04977dc68cf2396dae8d3c879e","fileExtension":"jar","fileModified":"2022-08-18T16:57:58Z","fileSize":700138,"groupId":"za.co.absa","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.1"},
{"artifactId":"spark-data-standardization_2.11","description":"131e8637b3daeffcce7b6f62d6b46aff583dd490","fileExtension":"jar","fileModified":"2022-08-11T09:18:36Z","fileSize":700180,"groupId":"za.co.absa","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.0"},
{"artifactId":"hyperdrive-trigger","description":"646ff09b9dab78665f5e90ec299a636f342b53ce","fileExtension":"war","fileModified":"2022-08-12T14:47:59Z","fileSize":169137143,"groupId":"za.co.absa","hasJavadoc":false,"hasSignature":true,"hasSources":true,"name":"hyperdrive-trigger","packaging":"war","recordType":"artifact_add","version":"0.5.12"},
{"artifactId":"web-sheaf","description":"e2a3a18cac6ed789d4c06994183fda2c4df83a8b","fileExtension":"jar","fileModified":"2022-08-25T17:07:58Z","fileSize":33259,"groupId":"zone.src.sheaf","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"${project.groupId}:${project.artifactId}","packaging":"jar","recordType":"artifact_add","version":"1.0.4"},
{"artifactId":"shared_2.12","description":"25ec9fec4ed4217d6fbb3156e48ad0eff9e614bf","fileExtension":"jar","fileModified":"2022-08-04T07:28:21Z","fileSize":13133,"groupId":"za.co.absa.hyperdrive","hasJavadoc":true,"hasSignature":true,"hasSources":true,"packaging":"jar","recordType":"artifact_add","version":"4.7.0"},
{"artifactId":"sheaf-parent","description":"925873572af8a4d9658bea17c88f3b0ab1566749","fileExtension":"pom","fileModified":"2022-08-25T15:53:20Z","fileSize":14995,"groupId":"zone.src.sheaf","hasJavadoc":false,"hasSignature":true,"hasSources":false,"name":"${project.groupId}:${project.artifactId} pom","packaging":"pom","recordType":"artifact_add","version":"1.11"},
{"artifactId":"sheaf-deps-bom","description":"264c0e81886d6e784db33498c4a5658b02a601d4","fileExtension":"pom","fileModified":"2022-08-25T16:05:27Z","fileSize":7075,"groupId":"zone.src.sheaf","hasJavadoc":false,"hasSignature":true,"hasSources":false,"name":"${project.groupId}:${project.artifactId}","packaging":"pom","recordType":"artifact_add","version":"1.3"},
```

## Why?
I know, I know...don't worry, I have my reasons :)
