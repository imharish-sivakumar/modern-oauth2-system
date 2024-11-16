# OAuth2 Local testing scrapper

## Pre requisite:

Copy over `.env.example` to `.env` and fill in the values for cisauth environmental variables.

Run
``
npm run proxy-cisauth
`` command to run the cisauth proxy

Run
``
npm run cisauth
``
command to run the cisauth

# Run below services

* Oauth2 Server
* UMS for cisauth
* TMS (make sure to copy and paste the client ids and secrets in the TMS config and env)
* Proxy Server (proxy.js)
* Replace your account details in the env
* Run the file to return access token and session ID

Note:
Make sure to run service in their respective ports mentioned below.

* UMS(REST)    : 8012
* AMS(REST)    : 8016
* TMS(gRPC): 5052
* Proxy Server: 3000
* OAuthAdmin: 4445
* OAuthPublic: 4444
* AdminOAuthAdmin: 4447
* OAuthPublic: 4446