import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Layout } from './components/Layout'
import { ModelCRUD } from './components/ModelCRUD'
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Truck, Cpu, Bell, Activity } from 'lucide-react'

function Dashboard() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Global Fleet Overview</h1>
        <p className="text-slate-500">FleetManager Enterprise IoT Platform</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Fleet Health</CardTitle>
            <Activity className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">98.2%</div>
            <p className="text-xs text-muted-foreground">Optimal operational status</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Devices</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">1,248</div>
            <p className="text-xs text-muted-foreground">+42 new since last login</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Alerts (24h)</CardTitle>
            <Bell className="h-4 w-4 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">14</div>
            <p className="text-xs text-muted-foreground">2 critical requires attention</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Fuel Efficiency</CardTitle>
            <Truck className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">8.4 L/100km</div>
            <p className="text-xs text-muted-foreground">-0.2% improvement</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="col-span-1">
          <CardHeader><CardTitle>Operational Status</CardTitle></CardHeader>
          <CardContent className="h-[200px] flex items-center justify-center border-t">
             <span className="text-slate-400">Live Telemetry Visualization Placeholder</span>
          </CardContent>
        </Card>
        <Card className="col-span-1">
          <CardHeader><CardTitle>Geofence Violations</CardTitle></CardHeader>
          <CardContent className="h-[200px] flex items-center justify-center border-t">
             <span className="text-slate-400">Map Visualization Placeholder</span>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/organizations" element={<ModelCRUD title="Organization" endpoint="organizations" />} />
          <Route path="/fleets" element={<ModelCRUD title="Fleet" endpoint="fleets" />} />
          <Route path="/devices" element={<ModelCRUD title="Device" endpoint="devices" />} />
          <Route path="/sensors" element={<ModelCRUD title="Sensor" endpoint="sensors" />} />
          <Route path="/telemetries" element={<ModelCRUD title="Telemetry" endpoint="telemetries" />} />
          <Route path="/drivers" element={<ModelCRUD title="Driver" endpoint="drivers" />} />
          <Route path="/assets" element={<ModelCRUD title="Asset" endpoint="assets" />} />
          <Route path="/trips" element={<ModelCRUD title="Trip" endpoint="trips" />} />
          <Route path="/geofences" element={<ModelCRUD title="Geofence" endpoint="geofences" />} />
          <Route path="/alerts" element={<ModelCRUD title="Alert" endpoint="alerts" />} />
          <Route path="/maintenance" element={<ModelCRUD title="Maintenance" endpoint="maintenance_tasks" />} />
        </Routes>
      </Layout>
    </Router>
  )
}

export default App
