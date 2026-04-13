# GoFrame Admin UI

Modern admin panel built with React + Vite + Tailwind CSS + shadcn/ui.

## Tech Stack

- **Bundler**: Vite
- **Framework**: React 19 (TypeScript)
- **Styling**: Tailwind CSS 3
- **Components**: shadcn/ui
- **State Management**: Zustand
- **Routing**: React Router 7
- **Charts**: Recharts
- **Icons**: Lucide React

## Development

```bash
# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build
```

## Project Structure

```
src/
├── app/                    # App configuration and routing
├── components/
│   ├── ui/                 # shadcn/ui components
│   └── layout/             # Layout components
├── features/               # Feature-based modules
│   ├── auth/               # Authentication
│   ├── overview/           # Dashboard overview
│   ├── data-studio/        # Data export/import
│   ├── system/             # System metrics
│   ├── network/            # Network inspector
│   ├── infra/              # Session management
│   ├── health/             # Health checks
│   ├── rbac/               # Access control
│   └── audit/              # Audit log
├── hooks/                  # Custom hooks
├── lib/                    # Utilities
├── services/               # API services
├── stores/                 # Zustand stores
├── types/                  # TypeScript types
└── main.tsx                # Entry point
```

## Features

- 🔐 **Authentication**: Login page with session management
- 📊 **Overview**: Dashboard with model statistics
- 💾 **Data Studio**: Export/Import data (CSV, JSON, SQL)
- 🖥️ **System Pulse**: Real-time Go runtime metrics
- 🌐 **Network Inspector**: Live HTTP traffic monitoring
- 👥 **Session Manager**: Active user session management
- ❤️ **Health Checks**: Service health monitoring
- 🛡️ **Access Control**: RBAC policy management
- 📝 **Audit Log**: Administrative action tracking

## Building for Production

The admin UI is automatically embedded into the Go binary. To rebuild:

```bash
# From the project root
./pkg/admin/build-ui.sh

# Or manually
cd pkg/admin/ui
npm install
npm run build
```

The built files will be placed in `ui/dist/` and embedded via Go's `//go:embed` directive.

## Customization

This admin panel is fully customizable. Modify:

- **Theme**: Edit `tailwind.config.js` and `src/index.css`
- **Components**: Modify `src/components/ui/` (shadcn/ui components)
- **Features**: Add new pages in `src/features/`
- **API**: Extend `src/services/api.ts`

## Architecture

Follows feature-based architecture:
- Each feature is self-contained with its own components, pages, and logic
- Global state managed via Zustand stores
- API calls centralized in services
- UI components use shadcn/ui for consistency
