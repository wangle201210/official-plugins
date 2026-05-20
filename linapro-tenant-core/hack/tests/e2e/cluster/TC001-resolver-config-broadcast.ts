import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0218 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-1 解析策略无广播', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-1a: removed resolver policy API does not create shared revision', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0218();
  });
});
