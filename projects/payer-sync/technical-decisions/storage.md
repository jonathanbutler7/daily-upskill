# Storage

## GCS vs Filestore, or object storage vs file systems

File systems tend to shine when you need to update, seek, and lock files, while an object store is useful when you need to write once and read many times, without need for updates.

A key structural difference is that file systems are organized in a hierarchichal tree structure while object storage is flat (key/value pairs).

A good use case for object storage would be a service that handles uploading profile photos. 

For this product, we will need to ingest from an SFTP server that may be organized like a file store with a hierarchichal tree structure, but we can store the ingested files in an object store.

## Database

We will need a relational database that supports multiple tables and foreign key references. Also need ACID transactions.

## DB size requirements

## Hot vs cold storage