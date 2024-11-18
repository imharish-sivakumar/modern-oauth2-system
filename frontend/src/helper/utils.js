import loginService from '../services/loginService';

export const broadCastFetchLoginChallenge = async (code) => {
    const loginChallengeResponse = await loginService.fetchLoginChallenge(code);
    const loginUrl = loginChallengeResponse.request.responseURL;
    const query = new URL(loginUrl);
    return query.searchParams.get('login_challenge');
};

const sha256 = (plain) => {
    const encoder = new TextEncoder();
    const data = encoder.encode(plain);
    return window.crypto.subtle.digest('SHA-256', data);
};

const base64urlencoded = (str) => {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(str)))
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/, '');
};

export const getCodeChallenge = async (code) => {
    const hashed = await sha256(code);
    return base64urlencoded(hashed);
};

export const generateRandomString = () => {
    const array = new Uint32Array(28);
    window.crypto.getRandomValues(array);
    return Array.from(array, (dec) => `0${dec.toString(16)}`.substr(-2)).join('');
};

export const handleConsentAndFetchToken = async (url, loginVerifier) => {
    const acceptLoginResponse = await loginService.loginConsentInitiate(url);
    const consentUrl = acceptLoginResponse.request.responseURL;
    const consentChallenge = new URL(consentUrl).searchParams.get('consent_challenge');
    const consentAcceptResponse = await loginService.loginConsentAccept(consentChallenge);
    const consentCallbackResponse = await loginService.loginConsentCallback(consentAcceptResponse.data.redirect_to);
    const loginCode = new URL(consentCallbackResponse.request.responseURL).searchParams.get('code');
    const exchangeTokenResponse = await loginService.loginExchangeToken(loginCode, loginVerifier);
    const {
        accessToken: accessToken,
        sessionID: sessionID,
        expiresIn: expiresIn,
        expiresAt: expiresAt,
    } = exchangeTokenResponse.data;
    return { accessToken, sessionID, expiresIn, expiresAt };
};
