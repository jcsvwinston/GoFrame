#!/bin/bash
# Build the admin UI for production embedding
set -e

echo "🔧 Building GoFrame Admin UI..."

cd "$(dirname "$0")"

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
  echo "📦 Installing dependencies..."
  npm install
fi

# Build for production
echo "🏗️  Building production bundle..."
npm run build

echo "✅ Admin UI built successfully!"
echo "📁 Output: ui/dist/"
