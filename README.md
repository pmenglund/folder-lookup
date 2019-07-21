# Folder Lookup

GCP Cloud Function that dumps a folder lookup table into BigQuery.

The code walks the folder structure and dumps a BigQuery table which can be used as a
lookup-table to get from folder id to folder name, which is useful when you create
DataStudio visualizations for billing information.


## Parameters

There are five environment variables

* `ROOT`  
  this is where to start the folder traverse, either `organizations/nnnnnn` or `folders/nnnnnn`
* `MAX_DEPTH`  
  this how far down you want to traverse, defaults to `4` which is the max depth of the folders available in GCP
* `DATASET`  
  name of the BigQuery dataset where to save the folder names
* `TABLE`  
  name of the BigQuery table in the above dataset where to save the folder names, defaults to `folders`
* `PROJECT`  
  name of the project where to run the BigQuery jobs that inserts/updates/deletes the table

## Required Setup

First you will need to enable Cloud Resource Manager API in the project.

### BigQuery

You will need to create the BigQuery dataset and table 

```json
[
    {
        "name": "id",
        "type": "STRING",
        "mode": "REQUIRED"
    },
    {
        "name": "name",
        "type": "STRING",
        "mode": "REQUIRED"
    },
    {
        "name": "level",
        "type": "INTEGER",
        "mode": "REQUIRED"
    },
    {
        "name": "parent",
        "type": "STRING",
        "mode": "REQUIRED"
    }
]
```

### Service Account

Create a service account and grant three different permissions to it.
This service which will be used for the Cloud Function.

1. Folder Viewer on the organization
1. BigQuery Job User on the project
1. BigQuery Data Editor on the dataset

### Source Repository

Either push the code in this repo to a new source repository,
or clone this one and use connect external repository.

### Cloud Function

Create a cloud function that uses the above repo and is triggered by pub/sub topic.

### Cloud Scheduler

Create a Cloud Scheduler job that triggers the above pub/sub tropic.

## Testing

You can run the Circle CI cli locally and test, with:

```bash
$ circleci local execute
```
  
