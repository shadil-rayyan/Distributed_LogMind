#!/usr/bin/env node

const baseUrl = process.env.BASE_URL || "http://localhost:8080";

const logs = [
  { service: "payment-api", level: "error", message: "Database connection failed" },
  { service: "payment-api", level: "error", message: "Timeout while querying users" },
  { service: "payment-api", level: "error", message: "Retry budget exhausted" },
];

async function sendLog(payload) {
  const response = await fetch(`${baseUrl}/logs`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const body = await response.text();
  console.log(`POST /logs -> ${response.status} ${body}`);
}

async function fetchIncidents() {
  const response = await fetch(`${baseUrl}/incidents`);
  const body = await response.text();
  console.log(`GET /incidents -> ${response.status} ${body}`);
}

for (const payload of logs) {
  await sendLog(payload);
}

await fetchIncidents();
