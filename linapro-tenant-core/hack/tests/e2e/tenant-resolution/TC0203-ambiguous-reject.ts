import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0203 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-203 固定 prompt 歧义策略', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-203a: resolver policy stays code-owned and rejects reject mode', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0203();
  });
});
