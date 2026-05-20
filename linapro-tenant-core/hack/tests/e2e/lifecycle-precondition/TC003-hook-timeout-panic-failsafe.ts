import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0212 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-3 钩子 fail-safe', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-3a: tenant delete does not rely on lifecycle event outbox', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0212();
  });
});
