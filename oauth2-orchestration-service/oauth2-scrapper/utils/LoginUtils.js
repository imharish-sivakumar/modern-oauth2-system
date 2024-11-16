import axios from 'axios';
import crypto from 'node:crypto';

const StatusOK = 200;
const StatusSeeOther = 303;
const StatusNotModified = 304;

const sha256 = plain => {
    const encoder = new TextEncoder();
    const data = encoder.encode(plain);
    return crypto.webcrypto.subtle.digest('SHA-256', data);
};

const base64urlencoded = str => {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(str)))
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/, '');
};

export const getCodeChallenge = async code => {
    const hashed = await sha256(code);
    return base64urlencoded(hashed);
};

export const generateRandomString = () => {
    const array = crypto.randomBytes(28);
    return Array.from(array, dec => `0${dec.toString(16)}`.slice(-2)).join('');
};

export const fetchLoginChallenge = async ({
    oauthURL,
    code,
    webClientID,
    oauthRedirectURI
}) => {
    const codeChallenge = await getCodeChallenge(code);

    return axios.get(oauthURL, {
        params: {
            response_type: 'code',
            client_id: webClientID,
            state: generateRandomString(),
            scope: 'offline_access openid',
            redirect_uri: oauthRedirectURI,
            code_challenge: codeChallenge,
            code_challenge_method: 'S256'
        },
        maxRedirects: 0,
        validateStatus(status) {
            return status >= StatusOK && status < StatusSeeOther; // default
        }
    });
};

export const handleConsentAndFetchToken = async ({
    url,
    loginVerifier,
    authenticationCSRFCookie,
    tokenExchangeURL,
    webClientID,
    oauthRedirectURI,
    oauthURL
}) => {
    const redirectURL = new URL(url);
    const acceptLoginResponse = await axios.get(oauthURL + redirectURL.search, {
        headers: {
            Cookie: authenticationCSRFCookie
        },
        maxRedirects: 0,
        validateStatus(status) {
            return status >= StatusOK && status < StatusSeeOther; // default
        }
    });
    const loginConsentURL = acceptLoginResponse.request.res.headers['location'];
    const consentCSRFCookie =
        acceptLoginResponse.request.res.headers['set-cookie'];
    const authenticationSession = consentCSRFCookie[0].split(' ')[0];
    const consentSession = consentCSRFCookie[1].split(' ')[0];
    const consentRedirection = await axios.get(loginConsentURL, {
        headers: {
            Cookie: authenticationSession + ' ' + consentSession
        }
    });
    const consentAcceptResponse = await axios.get(
        oauthURL + new URL(consentRedirection.data.redirect_to).search,
        {
            headers: {
                Cookie: authenticationSession + ' ' + consentSession
            },
            maxRedirects: 0,
            validateStatus(status) {
                return status >= StatusOK && status < StatusNotModified; // default
            }
        }
    );
    const loginCode = new URL(
        consentAcceptResponse.request.res.headers['location']
    ).searchParams.get('code');
    const exchangeTokenResponse = await loginExchangeToken({
        tokenExchangeURL,
        loginCode,
        webClientID,
        oauthRedirectURI,
        loginVerifier
    });
    const { accessToken, expiresIn, expiresAt } = exchangeTokenResponse.data;
    return {
        access_token: accessToken,
        expires_at: expiresAt,
        expires_in: expiresIn,
        session_id: exchangeTokenResponse.request.res.headers['set-cookie'][4]
            .split(' ')[0]
            .split('session=')[1]
            .replace(';', '')
    };
};

export const loginExchangeToken = async ({
    tokenExchangeURL,
    loginCode,
    webClientID,
    oauthRedirectURI,
    loginVerifier
}) => {
    return axios.post(tokenExchangeURL, {
        code: loginCode,
        redirectUri: oauthRedirectURI,
        clientID: webClientID,
        codeVerifier: loginVerifier
    });
};
