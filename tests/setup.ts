import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';

// Global test setup
beforeEach(async () => {
  // Clean up any test databases
  const testDbPath = path.join(os.tmpdir(), 'mindport-test-*.db');
  try {
    const files = await fs.readdir(os.tmpdir());
    for (const file of files) {
      if (file.startsWith('mindport-test-') && file.endsWith('.db')) {
        await fs.unlink(path.join(os.tmpdir(), file));
      }
    }
  } catch (error) {
    // Ignore cleanup errors
  }
});

afterAll(async () => {
  // Final cleanup
  const testDbPath = path.join(os.tmpdir(), 'mindport-test-*.db');
  try {
    const files = await fs.readdir(os.tmpdir());
    for (const file of files) {
      if (file.startsWith('mindport-test-') && file.endsWith('.db')) {
        await fs.unlink(path.join(os.tmpdir(), file));
      }
    }
  } catch (error) {
    // Ignore cleanup errors
  }
});