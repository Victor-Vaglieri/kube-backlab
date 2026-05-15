const http = require('http');
const { Pool } = require('pg');

const port = process.env.PORT || 3000;

const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_PORT,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME,
});

// Initialize database table
const initDb = async () => {
  try {
    const client = await pool.connect();
    await client.query(`
      CREATE TABLE IF NOT EXISTS tasks (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        completed BOOLEAN DEFAULT FALSE
      );
      CREATE TABLE IF NOT EXISTS persistence_test (
        id SERIAL PRIMARY KEY,
        counter INTEGER DEFAULT 0,
        last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      );
      INSERT INTO persistence_test (counter) SELECT 0 WHERE NOT EXISTS (SELECT 1 FROM persistence_test);
    `);
    client.release();
    console.log('Database initialized');
  } catch (err) {
    console.error('Error initializing database', err);
  }
};

initDb();

const server = http.createServer(async (req, res) => {
  const { method, url } = req;

  // Health Checks
  if (url === '/healthz/liveness') {
    res.statusCode = 200;
    return res.end('OK');
  }

  if (url === '/healthz/readiness') {
    try {
      const client = await pool.connect();
      await client.query('SELECT 1');
      client.release();
      res.statusCode = 200;
      return res.end('OK');
    } catch (err) {
      res.statusCode = 500;
      return res.end('Service Unavailable');
    }
  }

  // API Routes
  try {
    if (url === '/pvc-test' && method === 'GET') {
      const updateResult = await pool.query('UPDATE persistence_test SET counter = counter + 1, last_update = NOW() RETURNING *');
      res.statusCode = 200;
      res.setHeader('Content-Type', 'application/json');
      return res.end(JSON.stringify({
        message: "PVC Persistence Test",
        description: "This counter survives pod restarts because it is stored in PostgreSQL with a PersistentVolume.",
        data: updateResult.rows[0]
      }));
    }

    if (url === '/tasks' && method === 'GET') {
      const result = await pool.query('SELECT * FROM tasks ORDER BY id ASC');
      res.statusCode = 200;
      res.setHeader('Content-Type', 'application/json');
      return res.end(JSON.stringify(result.rows));
    }

    if (url === '/tasks' && method === 'POST') {
      let body = '';
      req.on('data', chunk => { body += chunk.toString(); });
      req.on('end', async () => {
        try {
          const { title } = JSON.parse(body);
          const result = await pool.query('INSERT INTO tasks (title) VALUES ($1) RETURNING *', [title]);
          res.statusCode = 201;
          res.setHeader('Content-Type', 'application/json');
          res.end(JSON.stringify(result.rows[0]));
        } catch (e) {
          res.statusCode = 400;
          res.end('Invalid JSON');
        }
      });
      return;
    }

    if (url === '/worker' && method === 'GET') {
      http.get('http://worker-service.dev.svc.cluster.local', (workerRes) => {
        let data = '';
        workerRes.on('data', (chunk) => { data += chunk; });
        workerRes.on('end', () => {
          res.statusCode = 200;
          res.setHeader('Content-Type', 'text/plain');
          res.end(`Response from Worker: ${data}`);
        });
      }).on('error', (err) => {
        res.statusCode = 500;
        res.end(`Worker unreachable: ${err.message}`);
      });
      return;
    }

    // Default route
    const client = await pool.connect();
    const result = await client.query('SELECT NOW() as time');
    client.release();
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end(`Kube-Backlab API v1. DB Time: ${result.rows[0].time}\nUse /tasks (GET/POST) for CRUD operations.\n`);
  } catch (err) {
    console.error('API Error', err);
    res.statusCode = 500;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Internal Server Error\n');
  }
});

server.listen(port, () => {
  console.log(`running at http://localhost:${port}/`);
});
