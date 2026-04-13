import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useToast } from '@/components/ui/use-toast'
import * as api from '@/services/api'
import { Upload, Download, FileText, Loader2 } from 'lucide-react'

export default function DataStudioPage() {
  const { toast } = useToast()
  const [exportFormat, setExportFormat] = useState<'csv' | 'json' | 'sql'>('json')
  const [exportModel, setExportModel] = useState('')
  const [isExporting, setIsExporting] = useState(false)
  const [isImporting, setIsImporting] = useState(false)
  const [importFile, setImportFile] = useState<File | null>(null)

  const handleExport = async () => {
    setIsExporting(true)
    try {
      const url = await api.exportData(exportFormat, exportModel || undefined)
      toast({
        title: 'Export successful',
        description: 'Your data has been exported',
      })
      window.open(url, '_blank')
    } catch (error) {
      toast({
        variant: 'destructive',
        title: 'Export failed',
        description: 'Failed to export data',
      })
    } finally {
      setIsExporting(false)
    }
  }

  const handleImport = async () => {
    if (!importFile) {
      toast({
        variant: 'destructive',
        title: 'No file selected',
        description: 'Please select a file to import',
      })
      return
    }

    setIsImporting(true)
    try {
      await api.importData(importFile)
      toast({
        title: 'Import successful',
        description: 'Your data has been imported',
      })
      setImportFile(null)
    } catch (error) {
      toast({
        variant: 'destructive',
        title: 'Import failed',
        description: 'Failed to import data',
      })
    } finally {
      setIsImporting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Data Studio</h1>
        <p className="text-muted-foreground">Export and import your application data</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="h-5 w-5" />
              Export Data
            </CardTitle>
            <CardDescription>Download your data in various formats</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="format">Format</Label>
              <div className="flex gap-2">
                {(['csv', 'json', 'sql'] as const).map((format) => (
                  <Button
                    key={format}
                    variant={exportFormat === format ? 'default' : 'outline'}
                    onClick={() => setExportFormat(format)}
                    className="flex-1"
                  >
                    {format.toUpperCase()}
                  </Button>
                ))}
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="model">Model (optional)</Label>
              <Input
                id="model"
                placeholder="Leave empty for all models"
                value={exportModel}
                onChange={(e) => setExportModel(e.target.value)}
              />
            </div>

            <Button
              onClick={handleExport}
              disabled={isExporting}
              className="w-full"
            >
              {isExporting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Exporting...
                </>
              ) : (
                <>
                  <Download className="mr-2 h-4 w-4" />
                  Export Data
                </>
              )}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Upload className="h-5 w-5" />
              Import Data
            </CardTitle>
            <CardDescription>Upload and import data files</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="file">File</Label>
              <div className="border-2 border-dashed rounded-lg p-6 text-center">
                <input
                  type="file"
                  id="file"
                  accept=".csv,.json,.sql"
                  onChange={(e) => setImportFile(e.target.files?.[0] || null)}
                  className="hidden"
                />
                <label
                  htmlFor="file"
                  className="cursor-pointer flex flex-col items-center gap-2"
                >
                  <FileText className="h-8 w-8 text-muted-foreground" />
                  <span className="text-sm">
                    {importFile ? importFile.name : 'Click to select a file'}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    CSV, JSON, or SQL files supported
                  </span>
                </label>
              </div>
            </div>

            <Button
              onClick={handleImport}
              disabled={isImporting || !importFile}
              className="w-full"
            >
              {isImporting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Importing...
                </>
              ) : (
                <>
                  <Upload className="mr-2 h-4 w-4" />
                  Import Data
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
