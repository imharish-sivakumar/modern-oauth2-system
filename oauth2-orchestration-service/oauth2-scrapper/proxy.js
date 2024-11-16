import express from 'express';
import { createProxyMiddleware } from 'http-proxy-middleware';

const isAdminServer = process.argv[2] === 'admin';

const app = express();

const routerProxyConfig = {
    '/user-service': 'http://localhost:8080',
};

const options = {
    target: `http://localhost:8080`,
    changeOrigin: true,
    cookieDomainRewrite: '',
    secure: false,
    router: routerProxyConfig
};

const authProxy = {
    target: `http://localhost:4444`,
    changeOrigin: true,
    cookieDomainRewrite: '',
    secure: false
};


app.use(createProxyMiddleware('/user-service', options));

app.use(createProxyMiddleware('/oauth2', authProxy));

app.use('*', (req, res, next) => {
    res.set('content-type', 'text/html');
    res.set('Access-Control-Allow-Origin', '*');
    res.set(
        'Access-Control-Allow-Methods',
        'GET,PUT,POST,DELETE,PATCH,OPTIONS'
    );
    res.set(
        'Access-Control-Allow-Headers',
        'X-Requested-With, content-type, Authorization, Set-Cookie'
    );
    next();
});

app.listen(3000, err => {
    if (err) throw err;
    console.log(
        `> Ready on http://localhost:3000`
    );
});
