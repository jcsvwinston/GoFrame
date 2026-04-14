import { useState, useMemo } from 'react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import type { ModelSummary, RuntimeInfo } from '@/types'
import { Search, ChevronDown, ChevronRight, Table2, Box } from 'lucide-react'

interface Props {
  models: ModelSummary[]
  runtime: RuntimeInfo | null
  selectedModel: string | null
  selectedDbAlias: string | undefined
  onSelectModel: (name: string, dbAlias?: string) => void
}

type ViewMode = 'all' | 'engine' | 'database'

export default function ModelSidebar({ models, runtime, selectedModel, selectedDbAlias, onSelectModel }: Props) {
  const [search, setSearch] = useState('')
  const [viewMode, setViewMode] = useState<ViewMode>('all')
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())
  const [dbFilter, setDbFilter] = useState<string | null>(null)

  const multiEngine = (runtime?.engines?.length ?? 0) > 1
  const multiDb = (runtime?.databases?.length ?? 0) > 1

  const filtered = useMemo(() => {
    let list = models
    if (search.trim()) {
      const q = search.toLowerCase()
      list = list.filter(
        (m) =>
          m.name.toLowerCase().includes(q) ||
          m.table.toLowerCase().includes(q) ||
          (m.plural && m.plural.toLowerCase().includes(q)),
      )
    }
    if (dbFilter) {
      list = list.filter((m) => m.databases?.includes(dbFilter))
    }
    return list
  }, [models, search, dbFilter])

  const engineGroups = useMemo(() => {
    if (!runtime?.engine_groups) return []
    return runtime.engine_groups.map((eg) => ({
      ...eg,
      models: filtered.filter((m) =>
        eg.databases.some((db) => m.databases?.includes(db.alias)),
      ),
    })).filter((eg) => eg.models.length > 0)
  }, [runtime, filtered])

  const databaseGroups = useMemo(() => {
    if (!runtime?.databases) return []
    return runtime.databases.map((db) => ({
      ...db,
      models: filtered.filter((m) => m.databases?.includes(db.alias)),
    })).filter((g) => g.models.length > 0)
  }, [runtime, filtered])

  const toggleGroup = (name: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev)
      if (next.has(name)) next.delete(name)
      else next.add(name)
      return next
    })
  }

  const switchViewMode = (mode: ViewMode) => {
    setViewMode(mode)
    if (mode === 'engine' && runtime?.engines) {
      setExpandedGroups(new Set(runtime.engines))
    } else if (mode === 'database' && runtime?.databases) {
      setExpandedGroups(new Set(runtime.databases.map((d) => d.alias)))
    }
  }

  const renderModelItem = (m: ModelSummary, contextDbAlias?: string) => {
    const isActive = selectedModel === m.name && (contextDbAlias === undefined ? true : selectedDbAlias === contextDbAlias)

    return (
      <button
        key={`${m.name}-${contextDbAlias ?? 'all'}`}
        onClick={() => onSelectModel(m.name, contextDbAlias)}
        className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
          isActive
            ? 'bg-primary text-primary-foreground'
            : 'hover:bg-muted text-foreground'
        }`}
      >
        <span className="flex items-center gap-2 min-w-0">
          <Table2 className={`h-3.5 w-3.5 flex-shrink-0 ${isActive ? 'opacity-80' : 'opacity-40'}`} />
          <span className="truncate font-medium">{m.plural || m.name}</span>
        </span>
        <span className={`block text-[10px] ml-5.5 mt-0.5 ${isActive ? 'text-primary-foreground/70' : 'text-muted-foreground'}`}>
          {m.table}
        </span>
      </button>
    )
  }

  const renderGroupSection = (
    key: string,
    label: string,
    subtitle: string,
    items: ModelSummary[],
    contextDbAlias?: string,
  ) => {
    const isExpanded = expandedGroups.has(key)
    return (
      <div key={key}>
        <button
          onClick={() => toggleGroup(key)}
          className="w-full flex items-center gap-1.5 px-2 py-1.5 text-xs font-medium text-muted-foreground hover:text-foreground"
        >
          {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
          <Box className="h-3 w-3" />
          <span className="truncate">{label}</span>
          {subtitle && <span className="text-[10px] opacity-60 truncate">{subtitle}</span>}
          <Badge variant="outline" className="text-[10px] ml-auto flex-shrink-0">
            {items.length}
          </Badge>
        </button>
        {isExpanded && (
          <div className="ml-3 space-y-0.5">
            {items.map((m) => renderModelItem(m, contextDbAlias))}
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <div className="p-3 border-b space-y-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Filter models..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8 h-9"
          />
        </div>

        {(multiEngine || multiDb) && (
          <div className="flex gap-1 flex-wrap">
            <button
              onClick={() => switchViewMode('all')}
              className={`px-2 py-0.5 rounded text-xs transition-colors ${
                viewMode === 'all' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            {multiEngine && (
              <button
                onClick={() => switchViewMode('engine')}
                className={`px-2 py-0.5 rounded text-xs transition-colors ${
                  viewMode === 'engine' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:text-foreground'
                }`}
              >
                By Engine
              </button>
            )}
            {multiDb && (
              <button
                onClick={() => switchViewMode('database')}
                className={`px-2 py-0.5 rounded text-xs transition-colors ${
                  viewMode === 'database' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:text-foreground'
                }`}
              >
                By Database
              </button>
            )}
          </div>
        )}

        {viewMode === 'all' && multiDb && runtime?.databases && (
          <div className="flex flex-wrap gap-1">
            <button
              onClick={() => setDbFilter(null)}
              className={`px-1.5 py-0 rounded text-[10px] leading-5 transition-colors ${
                !dbFilter ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:text-foreground'
              }`}
            >
              All DBs
            </button>
            {runtime.databases.map((db) => (
              <button
                key={db.alias}
                onClick={() => setDbFilter(dbFilter === db.alias ? null : db.alias)}
                className={`px-1.5 py-0 rounded text-[10px] leading-5 transition-colors ${
                  dbFilter === db.alias ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:text-foreground'
                }`}
              >
                {db.alias}
              </button>
            ))}
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-2 space-y-0.5">
        {viewMode === 'all' && (
          filtered.length === 0 ? (
            <p className="text-xs text-muted-foreground text-center py-6">
              {search || dbFilter ? 'No models match your filter' : 'No models registered'}
            </p>
          ) : filtered.map((m) => renderModelItem(m))
        )}

        {viewMode === 'engine' && (
          engineGroups.length === 0 ? (
            <p className="text-xs text-muted-foreground text-center py-6">No engines available</p>
          ) : engineGroups.map((eg) =>
            renderGroupSection(eg.name, eg.name, `${eg.databases.length} db`, eg.models),
          )
        )}

        {viewMode === 'database' && (
          databaseGroups.length === 0 ? (
            <p className="text-xs text-muted-foreground text-center py-6">No databases available</p>
          ) : databaseGroups.map((db) =>
            renderGroupSection(db.alias, db.alias, db.dialect || db.engine || '', db.models, db.alias),
          )
        )}
      </div>

      {runtime && (
        <div className="border-t px-3 py-2 text-[11px] text-muted-foreground space-y-0.5">
          <div className="flex justify-between">
            <span>Models</span>
            <span>{runtime.models_total}</span>
          </div>
          {runtime.databases.length > 0 && (
            <div className="flex justify-between">
              <span>Databases</span>
              <span>{runtime.databases.length} ({runtime.engines.join(', ')})</span>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
