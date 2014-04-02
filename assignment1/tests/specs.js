var request = require('request'),
    mocha   = require('mocha'),
    expect  = require('chai').expect;

describe('Calendar API service', function() {
    var baseUrl = 'http://localhost:3000';
    var cid, eid;

    it('lists no calendars when no calendars have been created', function(done) {
        request.get(baseUrl + '/calendars/', function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(200);
            expect(body).to.equal("{}");
            done();
        });
    });

    it('can create a new calendar', function(done) {
        request({
            method: 'POST',
            uri: baseUrl + '/calendars/',
            json: {
                name: 'First'
            }
        }, function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(201);
            expect(body.name).to.equal("First");
            done();
        });
    });

    it('lists all existing calendars', function(done) {
        request.get(baseUrl + '/calendars/', function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(200);
            var attrs = Object.keys(JSON.parse(body))
            cid = attrs[0];
            expect(attrs.length).to.not.equal(0);
            done();
        });
    });

    it('shows details for a specific calendar', function(done) {
        request.get(baseUrl + '/calendars/' + cid, function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(200);
            expect(JSON.parse(body).name).to.equal('First');
            done();
        });
    });

    it('it updates an existing calendar', function(done) {
        request({
            method: 'PUT',
            uri: baseUrl + '/calendars/' + cid,
            json: {
                name: 'Second'
            }
        }, function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(200);
            expect(body.name).to.equal('Second');
            done();
        });
    });

    it('it creates a calendar entry', function(done) {
        request({
            method: 'POST',
            uri: baseUrl + '/calendars/' + cid + '/entries/',
            json: {
                description: 'First entry',
                startTime: '2014-03-31 10:00:00 +0000',
                endTime: '2014-03-31 11:00:00 +0000'
            }
        }, function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(201);
            expect(body.description).to.equal('First entry');
            done();
        });
    });

     it('lists all existing calendar entries', function(done) {
        request.get(baseUrl + '/calendars/' + cid + '/entries/', function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(200);
            var attrs = Object.keys(JSON.parse(body))
            eid = attrs[0];
            expect(attrs.length).to.not.equal(0);
            done();
        });
    });

    it('it updates an existing calendar entry', function(done) {
        request({
            method: 'PUT',
            uri: baseUrl + '/calendars/' + cid + '/entries/' + eid,
            json: {
                startTime: '2014-03-31 11:00:00 +0000',
                endTime: '2014-03-31 12:00:00 +0000'
            }
        }, function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(200);
            expect(body.startTime).to.equal('2014-03-31 11:00:00 +0000');
            expect(body.endTime).to.equal('2014-03-31 12:00:00 +0000');
            done();
        });
    });

    it('it removes a calendar entry', function(done) {
        request({
            method: 'DELETE',
            uri: baseUrl + '/calendars/' + cid + '/entries/' + eid,
        }, function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(204);
            done();
        });
    });

    it('it removes a calendar', function(done) {
        request({
            method: 'DELETE',
            uri: baseUrl + '/calendars/' + cid,
        }, function(err, res, body) {
            expect(err).to.equal(null);
            expect(res.statusCode).to.equal(204);
            done();
        });
    });

});