/**
 * Created by Admin on 03.08.2015.
 */
'use strict';

var os = require('os'),
    pretty = require('prettysize');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/v2/status', applicationStatus);
    api.get('/api/v1/status', applicationStatus);
};

function applicationStatus(req, res) {
    if (application.Esl && !application.Esl['connecting']) {
        application.Esl.api('status', function(response) {
            res.json(getResult(response));
        });
    } else {
        res.json(getResult(false));
    };
};

function getResult (freeSwitchStatus) {
    return {
        "Version": process.env['VERSION'] || '',
        "Node memory": getMemoryUsage(),
        "Process ID": process.pid,
        "Process up time": formatTime(process.uptime()),
        "OS": getOsInfo(),
        "Users_Session": application.Users.length(),
        "Domain_Session": application.Domains.length(),
        "CRASH_WORKER_COUNT": process.env['CRASH_WORKER_COUNT'] || 0,
        "freeSWITCH": (freeSwitchStatus) ? freeSwitchStatus['body'] : 'Connect server error.'
    }
}

function getMemoryUsage () {
    var memory = process.memoryUsage();
    return {
        "rss": pretty(memory['rss']),
        "heapTotal": pretty(memory['heapTotal']),
        "heapUsed": pretty(memory['heapUsed'])
    }
};

function formatTime(seconds){
    function pad(s){
        return (s < 10 ? '0' : '') + s;
    }
    var hours = Math.floor(seconds / (60*60));
    var minutes = Math.floor(seconds % (60*60) / 60);
    var seconds = Math.floor(seconds % 60);

    return pad(hours) + ':' + pad(minutes) + ':' + pad(seconds);
};

function getOsInfo () {
    return {
        "Total memory": pretty(os.totalmem()),
        "Free memory": pretty(os.freemem()),
        "Platform": os.platform(),
        "Name": os.type(),
        "Architecture": os.arch()
    };
};

function getCpuInfo () {
    var res = {};
    var cpus = os.cpus();
    for(var i = 0, len = cpus.length; i < len; i++) {
        res['CPU' + i] = {};
        var cpu = cpus[i], total = 0;
        for(var type in cpu.times)
            total += cpu.times[type];

        for(type in cpu.times)
            res['CPU' + i][type] = Math.round(100 * cpu.times[type] / total)
    };
    return res;
};