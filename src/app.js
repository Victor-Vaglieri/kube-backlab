const http = require('http');
const { Pool } = require('pg');

const port = process.env.PORT || 3000;

// Configuração do Banco de Dados via Variáveis de Ambiente (Secret/ConfigMap)
const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_PORT,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME,
});

// Inicialização da Tabela para Teste de Persistência (PVC)
const initDb = async () => {
  try {
    const client = await pool.connect();
    await client.query(`
      CREATE TABLE IF NOT EXISTS demo_pvc (
        id SERIAL PRIMARY KEY,
        hits INTEGER DEFAULT 0,
        last_hit TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      );
      INSERT INTO demo_pvc (hits) SELECT 0 WHERE NOT EXISTS (SELECT 1 FROM demo_pvc);
    `);
    client.release();
    console.log('Database table "demo_pvc" initialized for PVC testing.');
  } catch (err) {
    console.error('Error initializing database:', err);
  }
};

initDb();

const server = http.createServer(async (req, res) => {
  const { method, url } = req;
  res.setHeader('Content-Type', 'application/json');

  // 1. Health Checks (Demonstração de Liveness/Readiness Probes)
  if (url === '/healthz/liveness') {
    res.statusCode = 200;
    return res.end(JSON.stringify({ status: 'ALIVE', message: 'Liveness probe OK' }));
  }

  if (url === '/healthz/readiness') {
    try {
      const client = await pool.connect();
      await client.query('SELECT 1'); // Testa a conexão real com o banco
      client.release();
      res.statusCode = 200;
      return res.end(JSON.stringify({ status: 'READY', message: 'Database connection OK' }));
    } catch (err) {
      res.statusCode = 503;
      return res.end(JSON.stringify({ status: 'NOT READY', error: 'Database unreachable' }));
    }
  }

  // 2. Teste de Persistência (Demonstração de PVC no Postgres)
  if (url === '/demo/pvc' && method === 'GET') {
    try {
      const updateResult = await pool.query('UPDATE demo_pvc SET hits = hits + 1, last_hit = NOW() RETURNING *');
      res.statusCode = 200;
      return res.end(JSON.stringify({
        feature: "Persistent Volume Claim (PVC)",
        description: "Este contador sobrevive ao reinício do Pod porque está armazenado no Postgres configurado com PVC.",
        data: updateResult.rows[0]
      }));
    } catch (err) {
      res.statusCode = 500;
      return res.end(JSON.stringify({ error: 'Database query failed' }));
    }
  }

  // 3. Teste de DNS Interno (Demonstração de comunicação Service-to-Service)
  if (url === '/demo/dns' && method === 'GET') {
    const target = 'http://worker-service.meu-projeto.svc.cluster.local';
    
    const request = http.get(target, (workerRes) => {
      let data = '';
      workerRes.on('data', (chunk) => { data += chunk; });
      workerRes.on('end', () => {
        res.statusCode = 200;
        res.end(JSON.stringify({
          feature: "Internal Kubernetes DNS",
          target: target,
          status: "SUCCESS",
          responseFromWorker: data.trim() || "Worker respondeu com sucesso (vazio)"
        }));
      });
    });

    request.on('error', (err) => {
      // Se for erro de Parse, mas a conexão foi feita, ainda prova que o DNS funciona!
      if (err.code === 'HPE_INVALID_CONSTANT' || err.message.includes('Parse Error')) {
        res.statusCode = 200;
        return res.end(JSON.stringify({
          feature: "Internal Kubernetes DNS",
          target: target,
          status: "SUCCESS (DNS OK)",
          message: "Conexão estabelecida via DNS, mas o worker não retornou um cabeçalho HTTP padrão.",
          rawError: err.message
        }));
      }
      
      res.statusCode = 500;
      res.end(JSON.stringify({ error: `Failed to reach worker service: ${err.message}` }));
    });
    return;
  }

  // Rota padrão
  res.statusCode = 200;
  res.end(JSON.stringify({
    message: "Kube-Backlab Demo App",
    endpoints: [
      "/demo/pvc - Testar persistência de dados",
      "/demo/dns - Testar comunicação com outro Pod via DNS",
      "/healthz/liveness - Liveness probe",
      "/healthz/readiness - Readiness probe"
    ]
  }));
});

server.listen(port, () => {
  console.log(`Demo app listening on port ${port}`);
});
