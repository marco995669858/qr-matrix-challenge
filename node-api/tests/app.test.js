const request = require('supertest');
const jwt = require('jsonwebtoken');
const { createApp } = require('../src/app');

const SECRET = 'test-secret';
const app = createApp({ jwtSecret: SECRET });
const token = jwt.sign({ sub: 'tester' }, SECRET, { algorithm: 'HS256', expiresIn: '5m' });

const body = {
  matrices: [
    { name: 'Q', data: [[1, 0], [0, 1]] },
    { name: 'R', data: [[2, 3], [0, 4]] },
  ],
};

describe('POST /api/stats', () => {
  test('200 con token válido', async () => {
    const res = await request(app)
      .post('/api/stats')
      .set('Authorization', `Bearer ${token}`)
      .send(body);

    expect(res.status).toBe(200);
    expect(res.body).toMatchObject({ max: 4, min: 0, sum: 11, anyDiagonal: true });
  });

  test('401 sin token', async () => {
    const res = await request(app).post('/api/stats').send(body);
    expect(res.status).toBe(401);
  });

  test('401 con token firmado con otro secreto', async () => {
    const bad = jwt.sign({ sub: 'x' }, 'otro-secreto');
    const res = await request(app)
      .post('/api/stats')
      .set('Authorization', `Bearer ${bad}`)
      .send(body);
    expect(res.status).toBe(401);
  });

  test('400 con matriz inválida', async () => {
    const res = await request(app)
      .post('/api/stats')
      .set('Authorization', `Bearer ${token}`)
      .send({ matrices: [{ name: 'A', data: [[1, 2], [3]] }] });
    expect(res.status).toBe(400);
    expect(res.body.error).toMatch(/rectangular/);
  });

  test('400 con JSON mal formado', async () => {
    const res = await request(app)
      .post('/api/stats')
      .set('Authorization', `Bearer ${token}`)
      .set('Content-Type', 'application/json')
      .send('{"matrices": [');
    expect(res.status).toBe(400);
  });
});

test('GET /health es público', async () => {
  const res = await request(app).get('/health');
  expect(res.status).toBe(200);
});
