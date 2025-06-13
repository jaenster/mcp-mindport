#!/usr/bin/env node

import { Command } from 'commander';
import { promises as fs } from 'fs';
import * as path from 'path';
import * as os from 'os';
import { SQLiteStorage } from './src/storage';
import { FuseSearch } from './src/search';
import { MCPServer } from './src/server';
import { Config } from './src/types';

const program = new Command();

program
  .name('mcp-mindport')
  .description('MindPort - MCP Resource Server with optimized search capabilities')
  .version('1.0.0');

program
  .option('-d, --daemon', 'run as daemon', false)
  .option('--domain <domain>', 'start server in specific domain context')
  .option('--create-domain', 'create domain if it doesn\'t exist', false)
  .option('--list-domains', 'list all available domains and exit', false)
  .option('--default-domain <domain>', 'set the default domain', 'default')
  .option('--log <path>', 'log file path')
  .option('--store-path <path>', 'storage directory path')
  .option('--search-path <path>', 'search index directory path')
  .option('--host <host>', 'server host for daemon mode', 'localhost')
  .option('--port <port>', 'server port for daemon mode', '8080')
  .action(async (options) => {
    try {
      await runServer(options);
    } catch (error) {
      console.error('Error:', error instanceof Error ? error.message : String(error));
      process.exit(1);
    }
  });

async function runServer(options: any) {
  // Set default paths
  const home = os.homedir();
  const storePath = options.storePath || 
    process.env.MCP_MINDPORT_STORE_PATH ||
    path.join(home, '.config', 'mindport', 'data', 'storage.db');
  
  const searchPath = options.searchPath || 
    process.env.MCP_MINDPORT_SEARCH_PATH ||
    path.join(home, '.config', 'mindport', 'data', 'search');

  // Handle environment variables
  const daemonMode = options.daemon || process.env.MCP_MINDPORT_DAEMON === 'true';
  const domain = options.domain || process.env.MCP_MINDPORT_DOMAIN || '';
  const defaultDomain = options.defaultDomain || process.env.MCP_MINDPORT_DEFAULT_DOMAIN || 'default';
  const logFile = options.log || process.env.MCP_MINDPORT_LOG || '';
  const host = options.host || process.env.MCP_MINDPORT_HOST || 'localhost';
  const port = parseInt(options.port || process.env.MCP_MINDPORT_PORT || '8080');

  // Configure logging
  if (logFile && logFile !== 'discard') {
    const logDir = path.dirname(logFile);
    await fs.mkdir(logDir, { recursive: true });
    
    // Redirect console.error to log file
    const logStream = await fs.open(logFile, 'a');
    const originalError = console.error;
    console.error = (...args) => {
      const timestamp = new Date().toISOString();
      logStream.write(`${timestamp} ${args.join(' ')}\n`);
    };
  } else if (!daemonMode) {
    // Disable logging in stdio mode to avoid interfering with MCP
    console.error = () => {};
  }

  // Create config
  const config: Config = {
    server: {
      host,
      port,
    },
    storage: {
      path: storePath,
    },
    search: {
      indexPath: searchPath,
    },
    domain: {
      defaultDomain,
      isolationMode: 'hierarchical',
      allowCrossDomain: true,
      currentDomain: domain || defaultDomain,
    },
  };

  // Ensure directories exist
  await fs.mkdir(path.dirname(storePath), { recursive: true });
  await fs.mkdir(searchPath, { recursive: true });

  console.error(`Default domain set to: ${defaultDomain}`);

  // Initialize storage
  const storage = new SQLiteStorage(storePath);
  await storage.initialize();

  // Initialize search
  const search = new FuseSearch();

  // Create MCP server
  const mcpServer = new MCPServer(storage, search, config);

  // Handle domain operations
  if (options.listDomains) {
    const domains = await storage.listDomains();
    console.log('Available domains:');
    domains.forEach(d => {
      console.log(`• ${d.name} (${d.resourceCount} resources)`);
      if (d.description) {
        console.log(`  ${d.description}`);
      }
    });
    await storage.close();
    return;
  }

  if (domain) {
    if (options.createDomain) {
      await storage.createDomain(domain, `Domain for ${domain} context`);
      console.error(`Created domain: ${domain}`);
    }
    config.domain.currentDomain = domain;
    console.error(`Starting MindPort in domain context: ${domain}`);
  } else {
    console.error(`Starting MindPort in default domain context: ${defaultDomain}`);
  }

  // Handle graceful shutdown
  process.on('SIGINT', async () => {
    console.error('Shutting down...');
    await mcpServer.close();
    process.exit(0);
  });

  process.on('SIGTERM', async () => {
    console.error('Shutting down...');
    await mcpServer.close();
    process.exit(0);
  });

  if (daemonMode) {
    console.error('Daemon mode not yet implemented');
    process.exit(1);
  } else {
    // Run MCP server in stdio mode
    await mcpServer.start();
  }
}

if (require.main === module) {
  program.parse();
}