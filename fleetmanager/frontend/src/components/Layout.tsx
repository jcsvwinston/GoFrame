import { Link, useLocation } from 'react-router-dom'
import { LayoutDashboard, Cpu, Bell, Activity, Layers, User, Package, Map, Grid, ClipboardList, Settings } from 'lucide-react'

export function Layout({ children }: { children: React.ReactNode }) {
  const location = useLocation()
  
  const menuItems = [
    { name: 'Dashboard', icon: LayoutDashboard, path: '/' },
    { name: 'Organizations', icon: Layers, path: '/organizations' },
    { name: 'Fleets', icon: Grid, path: '/fleets' },
    { name: 'Devices', icon: Cpu, path: '/devices' },
    { name: 'Sensors', icon: Activity, path: '/sensors' },
    { name: 'Telemetry', icon: Activity, path: '/telemetries' },
    { name: 'Drivers', icon: User, path: '/drivers' },
    { name: 'Assets', icon: Package, path: '/assets' },
    { name: 'Trips', icon: Map, path: '/trips' },
    { name: 'Geofences', icon: Map, path: '/geofences' },
    { name: 'Alerts', icon: Bell, path: '/alerts' },
    { name: 'Maintenance', icon: ClipboardList, path: '/maintenance' },
  ]

  return (
    <div className="flex min-h-screen bg-slate-50">
      {/* Sidebar */}
      <aside className="w-64 bg-white border-r flex flex-col">
        <div className="p-6 border-b">
          <h1 className="text-xl font-bold bg-primary text-white p-2 rounded text-center">FleetManager</h1>
        </div>
        <nav className="flex-1 overflow-y-auto p-4 space-y-1">
          {menuItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center gap-3 px-3 py-2 rounded-md transition-colors ${
                  isActive ? 'bg-primary text-white' : 'text-slate-600 hover:bg-slate-100'
                }`}
              >
                <Icon className="h-4 w-4" />
                <span className="text-sm font-medium">{item.name}</span>
              </Link>
            )
          })}
        </nav>
        <div className="p-4 border-t">
          <div className="flex items-center gap-3 px-3 py-2 text-slate-600">
            <Settings className="h-4 w-4" />
            <span className="text-sm font-medium">Settings</span>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 p-8 overflow-y-auto">
        <div className="max-w-6xl mx-auto">
          {children}
        </div>
      </main>
    </div>
  )
}
