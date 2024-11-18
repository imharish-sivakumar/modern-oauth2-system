import JSEncrypt from 'jsencrypt';

const encrypt = (valueToEncrypt) => {
    const jsEncrypt = new JSEncrypt();
    jsEncrypt.setPublicKey(atob(process.env.REACT_APP_PASSWORD_ENC_PUBLIC_KEY));
    return jsEncrypt.encrypt(valueToEncrypt);
};

export default encrypt;
