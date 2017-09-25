# CLI example

This application creates a `termbox` based CLI application to send messages using Layer.  You must have a JSON file containing client credentials to run this, in the following format:

```
{
  "application_id": "<APPLICATION_ID>",
  "provider_id": "<PROVIDER_ID>",
  "account_id": "<ACCOUNT_ID>",
  "key": {
    "id": "<KEY_ID>",
    "private": "-----BEGIN RSA PRIVATE KEY-----\n<BASE64_PRIVATE_KEY>\n-----END RSA PRIVATE KEY-----"
  }
}
```

The above information is obtained when creating an application in the Layer dashboard.

The application can be run by specifying a username and a key:

```
$ go run ./main.go -u user1 -f PATH_TO_KEY
$ go run ./main.go -u user2 -f PATH_TO_KEY
```
