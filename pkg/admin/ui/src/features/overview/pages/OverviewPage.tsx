import { useEffect, useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import * as api from '@/services/api'
import type { Model } from '@/types'
import { Database, Table, Shield, Activity } from 'lucide-react'

export default function OverviewPage() {
  const [models, setModels] = useState<Model[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchModels = async () => {
      try {
        const data = await api.getModels()
        setModels(data)
      } catch (error) {
        console.error('Failed to fetch models:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchModels()
  }, [])

  const stats = [
    { label: 'Total Models', value: models.length, icon: Table, color: 'text-blue-500' },
    { label: 'Database', value: 'Connected', icon: Database, color: 'text-green-500' },
    { label: 'RBAC', value: 'Enabled', icon: Shield, color: 'text-purple-500' },
    { label: 'Status', value: 'Active', icon: Activity, color: 'text-orange-500' },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Overview</h1>
        <p className="text-muted-foreground">Welcome to GoFrame Admin Panel</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <Card key={stat.label}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{stat.label}</CardTitle>
              <stat.icon className={`h-4 w-4 ${stat.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Models</CardTitle>
          <CardDescription>Registered database models</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
            </div>
          ) : models.length === 0 ? (
            <p className="text-center text-muted-foreground py-4">No models registered</p>
          ) : (
            <div className="grid gap-2 md:grid-cols-2 lg:grid-cols-3">
              {models.map((model) => (
                <div
                  key={model.name}
                  className="flex items-center gap-3 p-3 rounded-lg border border-border hover:bg-accent transition-colors"
                >
                  <Table className="h-5 w-5 text-muted-foreground" />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{model.name}</p>
                    <p className="text-xs text-muted-foreground truncate">{model.table}</p>
                  </div>
                  <Badge variant="secondary">{model.fields?.length || 0} fields</Badge>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common administrative tasks</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-2">
              <button className="p-3 text-left rounded-lg border border-border hover:bg-accent transition-colors">
                <p className="font-medium">Export Data</p>
                <p className="text-sm text-muted-foreground">Download database contents</p>
              </button>
              <button className="p-3 text-left rounded-lg border border-border hover:bg-accent transition-colors">
                <p className="font-medium">View Audit Log</p>
                <p className="text-sm text-muted-foreground">Track administrative actions</p>
              </button>
              <button className="p-3 text-left rounded-lg border border-border hover:bg-accent transition-colors">
                <p className="font-medium">Manage Access Control</p>
                <p className="text-sm text-muted-foreground">Configure RBAC policies</p>
              </button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Information</CardTitle>
            <CardDescription>GoFrame Admin Panel</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex justify-between items-center py-2 border-b border-border">
                <span className="text-sm">Framework</span>
                <Badge>GoFrame</Badge>
              </div>
              <div className="flex justify-between items-center py-2 border-b border-border">
                <span className="text-sm">Admin UI</span>
                <Badge variant="secondary">React + Vite</Badge>
              </div>
              <div className="flex justify-between items-center py-2 border-b border-border">
                <span className="text-sm">Styling</span>
                <Badge variant="outline">Tailwind CSS</Badge>
              </div>
              <div className="flex justify-between items-center py-2">
                <span className="text-sm">Components</span>
                <Badge variant="outline">shadcn/ui</Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
