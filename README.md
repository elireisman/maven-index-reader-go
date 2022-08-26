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
{"artifactId":"spark-data-standardization_2.12","classifier":"sources","description":"a4ff4fdcf01f5b15e1a58869595af97275cd9c12","fileExtension":"jar","fileModified":"2022-08-18T16:57:58Z","fileSize":85429,"groupId":"za.co.absa","hasJavadoc":false,"hasSignature":true,"hasSources":false,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.1"},
{"artifactId":"spark-data-standardization_2.12","classifier":"javadoc","description":"d1808d0d094b19c5495ae533ae4e62f30e8707da","fileExtension":"jar","fileModified":"2022-08-18T16:57:58Z","fileSize":1802147,"groupId":"za.co.absa","hasJavadoc":false,"hasSignature":true,"hasSources":false,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.1"},
{"artifactId":"spark-data-standardization_2.12","description":"8347df993825619463c27b15a8591a77938e8739","fileExtension":"jar","fileModified":"2022-08-11T09:18:35Z","fileSize":500386,"groupId":"za.co.absa","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.0"},
{"artifactId":"spark-data-standardization_2.12","classifier":"sources","description":"3d4a07e72dd8d533c1b49e9e2197775fd75712bf","fileExtension":"jar","fileModified":"2022-08-11T09:18:35Z","fileSize":85433,"groupId":"za.co.absa","hasJavadoc":false,"hasSignature":true,"hasSources":false,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.0"},
{"artifactId":"spark-data-standardization_2.12","classifier":"javadoc","description":"a59e299a376987dcf0df498cf82ebb9b7fa3b279","fileExtension":"jar","fileModified":"2022-08-11T09:18:36Z","fileSize":1804029,"groupId":"za.co.absa","hasJavadoc":false,"hasSignature":true,"hasSources":false,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.0"},
{"artifactId":"spark-data-standardization_2.11","description":"472fff2e9b3c8e04977dc68cf2396dae8d3c879e","fileExtension":"jar","fileModified":"2022-08-18T16:57:58Z","fileSize":700138,"groupId":"za.co.absa","hasJavadoc":true,"hasSignature":true,"hasSources":true,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.1"},
{"artifactId":"spark-data-standardization_2.11","classifier":"sources","description":"a4ff4fdcf01f5b15e1a58869595af97275cd9c12","fileExtension":"jar","fileModified":"2022-08-18T16:57:58Z","fileSize":85429,"groupId":"za.co.absa","hasJavadoc":false,"hasSignature":true,"hasSources":false,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.1"},
{"artifactId":"spark-data-standardization_2.11","classifier":"javadoc","description":"53a8d2649ded7a6512f6a56c3a52c6986c91514e","fileExtension":"jar","fileModified":"2022-08-18T16:57:58Z","fileSize":892164,"groupId":"za.co.absa","hasJavadoc":false,"hasSignature":true,"hasSources":false,"name":"spark-data-standardization","packaging":"jar","recordType":"artifact_add","version":"0.1.1"},
```

## Why?
I know, I know...don't worry, I have my reasons :)
