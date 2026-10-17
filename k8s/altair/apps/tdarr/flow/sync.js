const flow = require('/flow/flow.json');
const url = 'http://tdarr/api/v2/cruddb';

const call = async (data) => {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ data: { collection: 'FlowsJSONDB', ...data } }),
  });
  if (!res.ok) throw new Error(`${data.mode} failed: ${res.status} ${await res.text()}`);
  return res.text();
};

(async () => {
  const body = await call({ mode: 'getById', docID: flow._id });
  const existing = body ? JSON.parse(body) : null;
  if (!existing) {
    await call({ mode: 'insert', docID: flow._id, obj: flow });
    console.log(`inserted ${flow._id}`);
    return;
  }
  const merged = { ...existing, ...flow };
  if (JSON.stringify(merged) === JSON.stringify(existing)) {
    console.log(`${flow._id} up to date`);
    return;
  }
  await call({ mode: 'update', docID: flow._id, obj: merged });
  console.log(`updated ${flow._id}`);
})().catch((err) => {
  console.error(err);
  process.exit(1);
});
