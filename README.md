# sqlc-gen-java

Generate type-safe code from SQL

## Usage

```yaml
version: '2'
plugins:
  - name: java
    wasm:
      url: https://github.com/colesturza/sqlc-gen-java/releases/download/v0.1.0/sqlc-gen-java.wasm
      sha256: ...
sql:
  - schema: src/main/resources/authors/postgresql/schema.sql
    queries: src/main/resources/authors/postgresql/query.sql
    engine: postgresql
    codegen:
      - out: src/main/kotlin/com/example/authors/postgresql
        plugin: java
        options:
          package: com.example.authors.postgresql
```
