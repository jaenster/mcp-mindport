# MindPort Web Browser

A simple web interface for browsing and searching MindPort MCP data.

## Features

- **Dashboard** - Overview of resources, prompts, and domains with statistics
- **Resource Browser** - Search and filter resources by domain and content
- **Resource Detail** - View full resource content with metadata and tags
- **Prompt Templates** - Browse and test prompt templates with variable substitution
- **Domain Management** - View and navigate between different domains

## Getting Started

```bash
# Install dependencies
cd site
npm install

# Start development server
npm run dev

# Visit http://localhost:3001
```

## Architecture

- **Next.js 14** - React framework with App Router
- **TypeScript** - Type safety for better development
- **SQLite Integration** - Direct access to MindPort database
- **Server-Side API** - RESTful endpoints for data access
- **Simple Styling** - Inline styles for maximum simplicity

## API Endpoints

- `GET /api/stats` - Dashboard statistics
- `GET /api/domains` - List all domains
- `GET /api/resources` - List resources (with filtering)
- `GET /api/resources/[id]` - Get specific resource
- `GET /api/prompts` - List prompt templates
- `GET /api/prompts/[id]` - Get specific prompt template

## Configuration

The web interface automatically connects to your MindPort database at:
`~/.config/mindport/data/storage.db`

No additional configuration required!

## Development

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm run start

# Lint code
npm run lint
```

## Features Overview

### Dashboard
- Total counts for resources, prompts, and domains
- Recent resources with quick access
- Clean, card-based layout

### Resource Browser
- Search across all resource content
- Filter by domain
- Tag visualization
- Content preview with truncation
- Pagination support

### Resource Detail View
- Full content display with syntax highlighting
- Complete metadata (domain, type, dates, URI)
- Tag listing
- Character count
- Navigation breadcrumbs

### Prompt Templates
- Variable input interface
- Real-time template rendering
- Copy to clipboard functionality
- Variable highlighting
- Domain filtering

### Domain Management
- Visual domain cards
- Resource counts per domain
- Quick navigation to domain resources
- Creation date tracking

## Technical Details

- **Database**: Read-only SQLite access
- **Performance**: Server-side rendering for initial load
- **Responsive**: Works on desktop and mobile
- **Simple**: No complex state management or external dependencies
- **Fast**: Direct database queries with minimal overhead

## Future Enhancements

- Real-time updates via WebSocket
- Full-text search highlighting
- Resource editing capabilities
- Export functionality
- Theme customization
- Bulk operations

Built for simplicity and functionality! 🚀