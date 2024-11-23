import axios from 'axios';
import { generateRandomString, getCodeChallenge } from '../helper/utils';

export default {
    loginUser: (email, encryptedPassword, loginChallenge) => {
        const requestBody = {
            email,
            password: encryptedPassword,
            loginChallenge: loginChallenge,
        };
        return axios.post(`/api/user-service/v1/login`, requestBody);
    },

    registerUser: (email, encryptedPassword) => {
        const requestBody = {
            email,
            password: encryptedPassword,
            confirmPassword: encryptedPassword,
        };

        return axios.post(`/api/user-service/v1/register`, requestBody);
    },

    verifyAccount: (verificationCode) => {
        const url = `/api/user-service/v1/verify?code=${verificationCode}`;
        return axios.get(url);
    },


    loginConsentInitiate: (url) => {
        return axios.get(url);
    },

    loginConsentAccept: (consentChallenge) => {
        const url = `/api/user-service/v1/login/consent?consent_challenge=${consentChallenge}`;
        return axios.get(url);
    },

    loginConsentCallback: (url) => {
        return axios.get(url);
    },

    loginExchangeToken: (loginCode, loginVerifier) => {
        return axios.post('/api/user-service/v1/token/exchange', {
            code: loginCode,
            redirectURI: process.env.REACT_APP_OAUTH_REDIRECT_URI,
            clientID: process.env.REACT_APP_OAUTH_CLIENT_ID,
            codeVerifier: loginVerifier,
        });
    },

    async fetchLoginChallenge(code) {
        const codeChallenge = await getCodeChallenge(code);
        const url = new URL(process.env.REACT_APP_OAUTH_URL);

        url.pathname = `${url.pathname}oauth2/auth`;
        url.searchParams.set('response_type', 'code');
        url.searchParams.set('client_id', process.env.REACT_APP_OAUTH_CLIENT_ID);
        url.searchParams.set('state', generateRandomString());
        url.searchParams.set('scope', 'openid');
        url.searchParams.set('redirect_uri', process.env.REACT_APP_OAUTH_REDIRECT_URI);
        url.searchParams.set('code_challenge', codeChallenge);
        url.searchParams.set('code_challenge_method', 'S256');
        console.log(url);

        return axios.get(url.href);
    },
};

