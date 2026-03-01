import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import crypto from 'k6/crypto';

// Custom metrics
const errorRate = new Rate('error_rate');
const inquiryDuration = new Trend('inquiry_duration', true);
const paymentDuration = new Trend('payment_duration', true);
const statusDuration = new Trend('status_duration', true);

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const CLIENT_KEY = __ENV.CLIENT_KEY || 'MK-9921-X';
const CLIENT_SECRET = __ENV.CLIENT_SECRET || 'super-secret-key-123';
const QRIS_PAYLOAD = '00020101021126690021ID.CO.BANKMANDIRI.WWW01189360000801299399930211712993999340303UKE51440014ID.CO.QRIS.WWW0215ID10232756067300303UKE5204274153033605802ID5912MIvanStore6015JakartaTimur';

// Test options: target 1000 RPS
export const options = {
  scenarios: {
    load_test: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 300,
      stages: [
        { duration: '15s', target: 50 },     // Warm up
        { duration: '15s', target: 200 },    // Ramp up
        { duration: '30s', target: 500 },    // Ramp to 500 RPS
        { duration: '1m', target: 1000 },    // Target 1000 RPS
        { duration: '1m', target: 1000 },    // Sustain 1000 RPS
        { duration: '15s', target: 0 },      // Ramp down
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<3000'],           // 95% under 3s (laptop/Docker)
    error_rate: ['rate<0.30'],                   // Error rate < 30% (optimistic lock conflicts expected)
    inquiry_duration: ['p(95)<2000'],             // Inquiry 95% < 2s
    payment_duration: ['p(95)<3000'],             // Payment 95% < 3s
    status_duration: ['p(95)<2000'],              // Status 95% < 2s
  },
};

// Generate HMAC-SHA256 signature
function generateSignature(method, path, timestamp, body) {
  const payload = method + path + timestamp + body;
  const hmacDigest = crypto.hmac('sha256', CLIENT_SECRET, payload, 'hex');
  return hmacDigest;
}

// Get current ISO8601 timestamp
function getTimestamp() {
  return new Date().toISOString();
}

// Build headers with HMAC signature
function buildHeaders(method, path, body) {
  const timestamp = getTimestamp();
  const signature = generateSignature(method, path, timestamp, body || '');
  return {
    'Content-Type': 'application/json',
    'X-Client-Key': CLIENT_KEY,
    'X-Timestamp': timestamp,
    'X-Signature': signature,
  };
}

// Flow: Inquiry -> Payment -> Status Check
export default function () {
  // Step 1: QRIS Inquiry
  const inquiryPath = `/api/qris/inquiry/${QRIS_PAYLOAD}`;
  const inquiryHeaders = buildHeaders('GET', inquiryPath, '');

  const inquiryStart = Date.now();
  const inquiryRes = http.get(`${BASE_URL}${inquiryPath}`, {
    headers: inquiryHeaders,
  });
  inquiryDuration.add(Date.now() - inquiryStart);

  const inquirySuccess = check(inquiryRes, {
    'inquiry status is 200': (r) => r.status === 200,
    'inquiry has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'success' && body.data && body.data.inquiry_id;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!inquirySuccess);

  if (!inquirySuccess) {
    sleep(0.1);
    return;
  }

  const inquiryData = JSON.parse(inquiryRes.body);
  const inquiryId = inquiryData.data.inquiry_id;

  // Step 2: QRIS Payment
  const paymentPath = '/api/qris/payment';
  const paymentBody = JSON.stringify({
    inquiry_id: inquiryId,
    user_id: 'user_123',
    amount: 1, // Small amount to avoid balance exhaustion
    payment_method: 'balance',
    pincode: '123456',
  });
  const paymentHeaders = buildHeaders('POST', paymentPath, paymentBody);

  const paymentStart = Date.now();
  const paymentRes = http.post(`${BASE_URL}${paymentPath}`, paymentBody, {
    headers: paymentHeaders,
  });
  paymentDuration.add(Date.now() - paymentStart);

  // Payment may fail due to optimistic lock conflict (expected under high concurrency)
  const paymentSuccess = check(paymentRes, {
    'payment accepted': (r) => r.status === 200 || r.status === 409,
  });

  // Only count non-conflict errors
  const isConflict = paymentRes.status === 409;
  const isPaymentOk = paymentRes.status === 200;
  errorRate.add(!isPaymentOk && !isConflict);

  if (!isPaymentOk) {
    sleep(0.1);
    return;
  }

  const paymentData = JSON.parse(paymentRes.body);
  const transactionId = paymentData.data.transaction_id;

  // Step 3: Check Transaction Status
  const statusPath = `/api/transaction/status/${transactionId}`;
  const statusHeaders = buildHeaders('GET', statusPath, '');

  const statusStart = Date.now();
  const statusRes = http.get(`${BASE_URL}${statusPath}`, {
    headers: statusHeaders,
  });
  statusDuration.add(Date.now() - statusStart);

  const statusSuccess = check(statusRes, {
    'status check is 200': (r) => r.status === 200,
    'transaction is SUCCESS': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.status === 'SUCCESS';
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!statusSuccess);

  sleep(0.1);
}

// Export summary to JSON file
export function handleSummary(data) {
  return {
    'k6/summary.json': JSON.stringify(data, null, 2),
  };
}
