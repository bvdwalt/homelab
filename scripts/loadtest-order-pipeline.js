// k6 load test for order-pipeline's order-service /orders endpoint.
//
// Usage:
//   kubectl --context=altair -n order-pipeline port-forward svc/order-service 8090:80
//   k6 run scripts/loadtest-order-pipeline.js
//
// Override via env vars, e.g.:
//   k6 run -e VUS=20 -e DURATION=30s -e BASE_URL=http://localhost:8090 scripts/loadtest-order-pipeline.js
import http from 'k6/http';
import { check } from 'k6';

const baseURL = __ENV.BASE_URL || 'http://localhost:8090';
const vus = parseInt(__ENV.VUS || '50', 10);
const duration = __ENV.DURATION || '60s';

export const options = {
  scenarios: {
    burst: {
      executor: 'constant-vus',
      vus,
      duration,
    },
  },
};

const items = ['widget', 'gadget', 'gizmo'];

export default function () {
  const item = items[Math.floor(Math.random() * items.length)];
  const payload = JSON.stringify({ items: [{ item, qty: 1 }] });
  const res = http.post(`${baseURL}/orders`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'status is 202': (r) => r.status === 202 });
}
