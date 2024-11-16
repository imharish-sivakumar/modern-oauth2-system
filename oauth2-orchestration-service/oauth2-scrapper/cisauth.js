import axios from 'axios';
import 'dotenv/config';
import {
    fetchLoginChallenge,
    generateRandomString,
    handleConsentAndFetchToken
} from './utils/LoginUtils.js';
import { RsaService } from './utils/RsaService.js';

const registerUser = async (email, password) => {
    const rsaService = new RsaService(process.env.CISAUTH_UI_WEB_PUBLIC_KEY);
    const encryptedPassword = await rsaService.encrypt(password);
    await axios.post('http://localhost:3000/user-service/v1/users', {
        email,
        password: encryptedPassword,
        confirmPassword: encryptedPassword
    });
};

const login = async () => {
    const loginVerifier = generateRandomString();
    const loginChallengeResponse = await fetchLoginChallenge({
        oauthURL: process.env.CISAUTH_OAUTH_URL,
        code: loginVerifier,
        webClientID: process.env.CISAUTH_UI_WEB_CLIENT_ID,
        oauthRedirectURI: process.env.CISAUTH_OAUTH_REDIRECT_URI
    });

    const loginUrl = loginChallengeResponse.request.res.headers['location'];
    const csrfCookieValue =
        loginChallengeResponse.request.res.headers['set-cookie'];
    const query = new URL(loginUrl);
    const loginChallengeCode = {
        loginChallenge: query.searchParams.get('login_challenge'),
        cookie: csrfCookieValue[0]
    };

    const rsaService = new RsaService(process.env.CISAUTH_UI_WEB_PUBLIC_KEY);
    const encryptedPassword = await rsaService.encrypt(
        process.env.CISAUTH_USER_PASSWORD
    );

    const data = await axios.post(
        'http://localhost:3000/user-service/v1/login',
        {
            email: process.env.CISAUTH_USER_EMAIL,
            password: encryptedPassword,
            loginChallenge: loginChallengeCode.loginChallenge
        }
    );

    const res = await handleConsentAndFetchToken({
        url: data.data.redirect_to,
        loginVerifier,
        authenticationCSRFCookie: loginChallengeCode.cookie,
        tokenExchangeURL:
            'http://localhost:3000/user-service/v1/token/exchange',
        webClientID: process.env.CISAUTH_UI_WEB_CLIENT_ID,
        oauthRedirectURI: process.env.CISAUTH_OAUTH_REDIRECT_URI,
        oauthURL: process.env.CISAUTH_OAUTH_URL
    });

    // Print the token and session
    console.log(res);
};

const main = async () => {
    registerUser(process.env.CISAUTH_USER_EMAIL, process.env.CISAUTH_USER_PASSWORD)
        .then(({data}) => {
            login();
        })
        .catch(err => {
            if (err.response.data.status === 'please try again') {
                login();
                return;
            }
            console.log('unable to create user ', err.response.data);
        });
};

main();
