import { ClientProxyFactory, Transport } from '@nestjs/microservices';
import { RABBITMQ_CONFIG } from '../src/config/rabbitmq.config';

async function testEmailService() {
  // Create client to emit events to email queue
  const client = ClientProxyFactory.create({
    transport: Transport.RMQ,
    options: {
      urls: RABBITMQ_CONFIG.urls,
      queue: RABBITMQ_CONFIG.queues.email,
    },
  });

  try {
    console.log('🧪 Testing Email Service Event Processing...\n');

    // Test 1: Send email notification event
    console.log('1️⃣ Testing: Emit notification.email event');
    const testData = {
      request_id: `test-${Date.now()}`,
      user: {
        id: '00000000-0000-0000-0000-000000000001',
        name: 'Test User',
        email: 'test.user@example.com', // Change this to your test email
        push_token: null,
      },
      template: {
        id: '00000000-0000-0000-0000-000000000002',
        name: 'welcome-email',
        type: 'email',
        subject: 'Welcome {{name}}!',
        body: '<h1>Hello {{name}}!</h1><p>Your email is {{email}}.</p>',
        variables: ['name', 'email'],
      },
      variables: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    // Emit event (fire and forget)
    client.emit('notification.email', testData);
    console.log('✅ Event emitted:', JSON.stringify(testData, null, 2));
    console.log('   Check Email Service logs to see if it was processed\n');

    // Wait a bit for processing
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // Test 2: Test idempotency (send same request_id twice)
    console.log('2️⃣ Testing: Idempotency (duplicate request_id)');
    const duplicateData = { ...testData, request_id: testData.request_id };
    client.emit('notification.email', duplicateData);
    console.log('✅ Duplicate event emitted with same request_id');
    console.log(
      '   Email Service should skip processing (already processed)\n',
    );

    // Test 3: Test with missing variables
    console.log('3️⃣ Testing: Email with missing variables');
    const missingVarsData = {
      request_id: `test-missing-${Date.now()}`,
      user: {
        id: '00000000-0000-0000-0000-000000000001',
        name: 'Test User',
        email: 'test.user2@example.com',
        push_token: null,
      },
      template: {
        id: '00000000-0000-0000-0000-000000000002',
        name: 'welcome-email',
        type: 'email',
        subject: 'Welcome {{name}}!',
        body: '<p>Hello {{name}}, your email is {{email}}.</p>',
        variables: ['name', 'email'],
      },
      variables: {
        name: 'Jane Doe',
        // email missing
      },
    };
    client.emit('notification.email', missingVarsData);
    console.log('✅ Event emitted with missing variables');
    console.log('   Variables should be replaced with empty strings\n');

    // Wait for processing
    await new Promise((resolve) => setTimeout(resolve, 2000));

    console.log('📊 Test Summary:');
    console.log('='.repeat(50));
    console.log('1. Email notification event emitted ✅');
    console.log('2. Idempotency test (duplicate request_id) ✅');
    console.log('3. Missing variables test ✅');
    console.log('='.repeat(50));
    console.log('\n📝 Next Steps:');
    console.log('1. Check Email Service console logs');
    console.log('2. Check email_logs table in database');
    console.log('3. Verify email was sent (check inbox/spam)');
    console.log('\n⚠️  Note: Email sending requires valid SMTP credentials.');
    console.log('   Set EMAIL_USER and EMAIL_PASS environment variables.');
  } catch (error) {
    console.error('❌ Error:', error);
    console.error('Error details:', JSON.stringify(error, null, 2));
  } finally {
    await client.close();
    console.log('\n✅ Test completed. Connection closed.');
    process.exit(0);
  }
}

testEmailService();
