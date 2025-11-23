import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
    scenarios: {
        team_scenario: {
            executor: 'constant-arrival-rate',
            rate: 250,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 70,
            maxVUs: 150,
            exec: 'teamWorkflowTest',
        },
        pr_scenario: {
            executor: 'constant-arrival-rate',
            rate: 250,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 70,
            maxVUs: 150,
            exec: 'prWorkflowTest',
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.0001'],
        http_req_duration: ['p(99)<100'],
    },
};

const BASE_URL = 'http://pr-reviewers-service:8080/api/v1';

let teams = [];
let prIds = [];

function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

function getAuthHeaders(token) {
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
    };
}

export function setup() {
    const loginRes = http.post(`${BASE_URL}/dummyLogin`, JSON.stringify({}), {
        headers: { 'Content-Type': 'application/json' },
    });

    const token = loginRes.json().token;

    for (let i = 0; i < 10; i++) {
        const teamName = `init_team_${i}_${Date.now()}`;
        const members = [
            { user_id: generateUUID(), username: `user1_${randomString(6)}` },
            { user_id: generateUUID(), username: `user2_${randomString(6)}` },
        ];

        const teamRes = http.post(`${BASE_URL}/team/add`, JSON.stringify({
            team_name: teamName,
            members: members
        }), {
            headers: getAuthHeaders(token),
            timeout: '10s'
        });

        if (teamRes.status === 201 || teamRes.status === 304) {
            teams.push({
                teamName: teamName,
                userIds: members.map(m => m.user_id)
            });
        }
    }

    return { token };
}

export function teamWorkflowTest(data) {
    const headers = getAuthHeaders(data.token);

    const teamName = `team_${Date.now()}_${randomString(8)}`;
    const members = [
        { user_id: generateUUID(), username: `user1_${randomString(6)}` },
        { user_id: generateUUID(), username: `user2_${randomString(6)}` },
    ];

    const createTeamRes = http.post(`${BASE_URL}/team/add`, JSON.stringify({
        team_name: teamName,
        members: members
    }), {
        headers,
        timeout: '5s'
    });

    check(createTeamRes, {
        'create team ok': (r) => r.status === 201 || r.status === 304
    });

    const getTeamRes = http.get(`${BASE_URL}/team/get?team_name=${encodeURIComponent(teamName)}`, {
        headers,
        timeout: '5s'
    });
    check(getTeamRes, {
        'get team info ok': (r) => r.status === 200
    });

    if (createTeamRes.status === 201) {
        teams.push({
            teamName: teamName,
            userIds: members.map(m => m.user_id)
        });
    }

    sleep(0.1);
}

export function prWorkflowTest(data) {
    const headers = getAuthHeaders(data.token);

    if (teams.length === 0) {
        const teamName = `temp_team_${Date.now()}_${randomString(8)}`;
        const members = [
            { user_id: generateUUID(), username: `user1_${randomString(6)}` },
            { user_id: generateUUID(), username: `user2_${randomString(6)}` },
        ];

        const createTeamRes = http.post(`${BASE_URL}/team/add`, JSON.stringify({
            team_name: teamName,
            members: members
        }), {
            headers,
            timeout: '5s'
        });

        if (createTeamRes.status === 201) {
            teams.push({
                teamName: teamName,
                userIds: members.map(m => m.user_id)
            });
        } else {
            return;
        }
    }

    const randomTeam = randomItem(teams);
    const authorId = randomItem(randomTeam.userIds);
    const prId = generateUUID();
    const prName = `PR_${randomString(10)}`;

    const createPRRes = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
        author_id: authorId,
        pull_request_id: prId,
        pull_request_name: prName
    }), {
        headers,
        timeout: '5s'
    });

    check(createPRRes, {
        'create PR ok': (r) => r.status === 201
    });

    const mergeRes = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify({
        pull_request_id: prId
    }), {
        headers,
        timeout: '5s'
    });

    check(mergeRes, {
        'merge PR ok': (r) => r.status === 200
    });

    sleep(0.1);
}
