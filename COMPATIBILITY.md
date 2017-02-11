Key:

- ✅ Yes : This feature is fully supported
- ✅ Emulated : This feature is not supported by the backend, but is fully emulated.
- ☑️ Partial : This feature is partially supported
- ⁿ̷ₐ Not Applicable : This feature is supported, and doesn't make sense to emulate.
- ⌛ Not Yet : This feature has not been implemented, but plans are to add it.
- ❌ No : This feature is supported by the backend, but there are no plans to add support to Kivik
- ？ Unknown : This feature has not been implemented, and no determination has yet been made on the feasibility of implementing it.

| API endpoint                                               | Implemented? | Kivik method(s) | CouchDB Driver | PouchDB Driver     | Memory Driver | Notes
| --------------------------------------------------------|--------------|-----------------|----------------|--------------------|---------------|---
| GET /                                                      | ☑️ partial   | Version()       | ✅ Yes          | ✅ Emulated        | ✅ Yes
| GET /_active_tasks                                         | ⌛ Not Yet   |                 | ⌛ Not Yet      | ⌛ Not Yet         | ⌛ Not Yet
| GET /_all_dbs                                              | ✅ Yes       | AllDBs()        | ✅ Yes          | ✅ Yes (w/ plugin) | ✅ Yes        | Unit tests broken in PouchDB due to an [apparent bug](https://github.com/nolanlawson/pouchdb-all-dbs/issues/25) in the pouchdb-all-dbs plugin.
| GET /_db_updates                                           | ⌛ Not Yet   |                 | ⌛ Not Yet      | ⌛ Not Yet         |
| GET /_log                                                  | ✅ Yes       | Log()           | ✅ Yes          | ⁿ̷ₐ Not Applicable  | ⌛ Not Yet
| GET /_replicate                                            | ⌛ Not Yet   |                 | ⌛ Not Yet      | ⌛ Not Yet         | ⌛ Not Yet
| GET /_restart                                              | ⌛ Not Yet   |                 | ⌛ Not Yet      | ⁿ̷ₐ Not Applicable  | ⌛ Not Yet
| GET /_stats                                                | ⌛ Not Yet   |                 | ⌛ Not Yet      | ？ Unknown         | ？ Unknown
| GET /_utils                                                | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⁿ̷ₐ Not Applicable  | ？ Unknown
| GET /_membership                                           | ✅ Yes       | Membership()    | ✅ Yes (2.0+ only) | ⁿ̷ₐ Not Applicable | ⁿ̷ₐ Not Applicable
| GET /favicon.ico                                           | ⁿ̷ₐ Not Applicable|             | ❌ No            | ⁿ̷ₐ Not Applicable | ⁿ̷ₐ Not Applicable
| POST /_session                                             | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⁿ̷ₐ Not Applicable |  ⁿ̷ₐ Not Applicable
| GET /_session                                              | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⁿ̷ₐ Not Applicable |  ⁿ̷ₐ Not Applicable
| DELETE /_session                                           | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⁿ̷ₐ Not Applicable |  ⁿ̷ₐ Not Applicable
| GET /_config                                               | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ？ Unknown       | ？ Unknown
| GET /_config/{section}                                     | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ？ Unknown       | ？ Unknown
| GET /_config/{section}/{key}                               | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ？ Unknown       | ？ Unknown
| PUT /_config/{section}/{key}                               | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ？ Unknown       | ？ Unknown
| DELETE /_config/{section}/{key}                            | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ？ Unknown       | ？ Unknown
| HEAD /{db}                                                 | ✅ Yes       | DBExists()      | ⌛ Not Yet      |  ⌛ Not Yet       | ✅ Yes
| GET /{db}                                                  | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}                                                  | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| DELETE /{db}                                               | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}                                                 | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_all_docs                                        | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_all_docs                                       | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_bulk_docs                                      | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_changes                                         | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_changes                                        | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_compact                                        | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_compact/{ddoc}                                 | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_ensure_full_commit                             | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_view_cleanup                                   | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_security                                        | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}/_security                                        | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_temp_view                                      | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ？ Unknown       | ？ Unknown
| POST /{db}/_purge                                          | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_missing_revs                                   | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_revs_diff                                      | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_revs_limit                                      | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}/_revs_limit                                      | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| HEAD /{db}/{docid}                                         | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/{docid}                                          | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}/{docid}                                          | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| DELETE /{db}/{docid}                                       | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| COPY /{db}/{docid}                                         | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| HEAD /{db}/{docid}/{attname}                               | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/{docid}/{attname}                                | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}/{docid}/{attname}                                | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| DELETE /{db}/{docid}/{attname}                             | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| HEAD /{db}/_design/{ddoc}                                  | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_design/{ddoc}                                   | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}/_design/{ddoc}                                   | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| DELETE /{db}/_design/{ddoc}                                | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| COPY /{db}/_design/{ddoc}                                  | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| HEAD /{db}/_design/{ddoc}/{attname}                        | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_design/{ddoc}/{attname}                         | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}/_design/{ddoc}/{attname}                         | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| DELETE /{db}/_design/{ddoc}/{attname}                      | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_design/{ddoc}/_info                             | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_design/{ddoc}/_view/{view}                      | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_design/{ddoc}/_view/{view}                     | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_design/{ddoc}/_show/{func}                      | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_design/{ddoc}/_show/{func}                     | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_design/{ddoc}/_show/{func}/{docid}              | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_design/{ddoc}/_show/{func}/{docid}             | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_design/{ddoc}/_list/{func}/{view}               | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_design/{ddoc}/_list/{func}/{view}              | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_design/{ddoc}/_list/{func}/{other-ddoc}/{view}  | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_design/{ddoc}/_list/{func}/{other-ddoc}/{view} | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| POST /{db}/_design/{ddoc}/_update/{func}                   | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}/_design/{ddoc}/_update/{func}/{docid}            | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| ANY /{db}/_design/{ddoc}/_rewrite/{path}                   | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| GET /{db}/_local/{docid}                                   | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| PUT /{db}/_local/{docid}                                   | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| DELETE /{db}/_local/{docid}                                | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
| COPY /{db}/_local/{docid}                                  | ⌛ Not Yet   |                 | ⌛ Not Yet      |  ⌛ Not Yet       | ⌛ Not Yet
