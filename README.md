# Layer Go Client
*NOTICE: This is an early release of the Layer Go client.  All API interfaces are subject to change.*

## Important Notes
* *This is an early release of the Layer Go client.  All API interfaces are subject to change.*
* Use of this API requires a Layer account
* Use of the Server API requires a bearer token, which can be obtained from the client dashboard at https://dashboard.layer.com
* Use of the Client API requires implementation of an external identity and authorization provider, or a credentials file (only recommended for testing)

## Using the Client API
### Creating a new client
```
ctx := context.Background()
return NewClient(ctx, "APP_ID", option.WithCredentials(&common.ClientCredentials{User: "USERNAME"}), option.WithTokenFunc(func(user, nonce string) (token string, err error) {
	// Make an HTTP call or perform local logic to create a signed JWT
	// with your private key.
	//
	// More details on this process can be found at:
	//   https://docs.layer.com/reference/client_api/authentication.out
	return
}))
```

## Using the Server API
### Creating a new client
```
ctx := context.Background()
client, err := NewClient(ctx, "APPLICATION_ID", option.WithBearerToken("API_KEY"))
if err != nil {
	fmt.Println(fmt.Sprintf("Error creating Layer Server API client: %v", err))
}
```
