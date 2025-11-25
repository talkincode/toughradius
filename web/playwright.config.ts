import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright 配置文件
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './e2e',
  
  /* 并行运行测试 */
  fullyParallel: true,
  
  /* 在 CI 上失败时重试 */
  retries: process.env.CI ? 2 : 0,
  
  /* CI 上使用所有 CPU，本地使用一半 */
  workers: process.env.CI ? 1 : undefined,
  
  /* Reporter */
  reporter: 'html',
  
  /* 共享设置 */
  use: {
    /* 基础 URL */
    baseURL: 'http://localhost:3000',
    
    /* 截图设置 */
    screenshot: 'only-on-failure',
    
    /* 视频设置 */
    video: 'retain-on-failure',
    
    /* 追踪设置 */
    trace: 'on-first-retry',
  },

  /* 配置项目 */
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },

    /* 移动端测试 */
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],

  /* 启动开发服务器 */
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});
