/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

var jwt = require('jwt-simple'),
    config = require(__appRoot + '/conf'),
    authService = require(__appRoot + '/services/auth'),
    tokenSecretKey = config.get('application:auth:tokenSecretKey');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.all('/api/v1/*', validateRequestV1);
    api.all('/api/v2/*', validateRequestV2);
    api.post('/login', login);
    api.post('/logout', logout);
};

function login (req, res, next) {
    var username = req.body.username || '';
    var password = req.body.password || '';

    if (username == '') {
        res.status(401);
        res.json({
            "status": 401,
            "message": "Invalid credentials"
        });
        return;
    };

    authService.login({
        username: username,
        password: password
    }, function (err, result) {
        if (err) {
            return next(err);
        };

        if (result) {
            return res
                .json(result);
        };

        return res
            .status(500)
            .json({
                "status": "error"
            });
    });
};

function logout (req, res, next) {
    try {
        var key = (req.body && req.body.x_key) || (req.query && req.query.x_key) || req.headers['x-key'];
        var token = (req.body && req.body.access_token) || (req.query && req.query.access_token) || req.headers['x-access-token'];
        if (!key || !token) {
            res.status(401);
            res.json({
                "status": 401,
                "message": "Invalid credentials"
            });
            return;
        };

        authService.logout({
            key: key,
            token: token
        }, function (err) {
            if (err) {
                return next(err);
            };

            res.status(200).json({
                "status": "OK",
                "info": "Successful logout."
            });
        });
    } catch (e) {
        next(e);
    }
};

function validateRequestV1(req, res, next) {
    try {
        var header = req.headers['authorization'] || '',
            token = header.split(/\s+/).pop() || '',
            auth = new Buffer(token, 'base64').toString(),
            parts = auth.split(/:/),
            username = parts[0],
            password = parts[1];

        return authService.baseAuth({
            "username": username,
            "password": password
        }, next);

    } catch (err) {
        res.status(500);
        return res.json({
            "status": 500,
            "message": "Oops something went wrong",
            "error": err
        });
    }
};

function validateRequestV2(req, res, next) {
    var token = (req.body && req.body.access_token) || (req.query && req.query.access_token) || req.headers['x-access-token'];
    var key = (req.body && req.body.x_key) || (req.query && req.query.x_key) || req.headers['x-key'];

    if (token && key) {
        try {
            var decoded = jwt.decode(token, tokenSecretKey);

            if (decoded.exp <= Date.now()) {
                return res
                    .status(400)
                    .json({
                        "status": 400,
                        "message": "Token Expired"
                    });
            };

            // Authorize the user to see if s/he can access our resources

            authService.validateUser(key, function (err, dbUser) {
                if (dbUser && dbUser.token == token) {
                    req['webitelUser'] = {
                        id: dbUser['username'],
                        domain: dbUser['domain'],
                        role: dbUser['role'],
                        roleName: dbUser['roleName'],
                        //testLeak: new Array(1e6).join('X')
                    };
                    next(); // To move to next middleware
                } else {
                    // No user with this name exists, respond back with a 401
                    return res
                        .status(401)
                        .json({
                            "status": 401,
                            "message": "Invalid User"
                        });
                }
            });

        } catch (err) {
            return res
                .status(500)
                .json({
                    "status": 500,
                    "message": "Oops something went wrong",
                    "error": err
                });
        };
    } else {
        return res
            .status(401)
            .json({
                "status": 401,
                "message": "Invalid Token or Key"
            });
    };
};