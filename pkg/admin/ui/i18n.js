(function() {
  "use strict";
  
  // Translation dictionaries
  var translations = {
    en: {
      "nav.overview": "Overview",
      "nav.data_studio": "Data Studio",
      "nav.system_pulse": "System Pulse",
      "nav.network_inspector": "Network Inspector",
      "nav.infra_manager": "Infra Manager",
      "nav.health": "Health",
      "nav.access_control": "Access Control",
      "nav.audit_log": "Audit Log",
      "nav.migrations": "Migrations",
      "nav.deployment": "Deployment",
      "nav.jobs": "Jobs",
      "nav.cache": "Cache",
      "nav.storage": "Storage",
      "nav.sites": "Sites",
      "cmd.search": "Search models, engines or actions",
      "theme.dark": "Dark",
      "theme.light": "Light",
      "btn.refresh": "Refresh",
      "btn.new_record": "New record",
      "btn.export_selected": "Export selected",
      "btn.export_all": "Export",
      "btn.import": "Import",
      "btn.delete_selected": "Delete selected",
      "btn.csv": "CSV",
      "runtime.all_systems": "All systems nominal",
      "export.title": "Export Data",
      "export.format": "Format",
      "export.now": "Export Now",
      "export.cancel": "Cancel",
      "export.complete": "Export complete: {records} records",
      "import.title": "Import Data",
      "import.file": "File (CSV or JSON)",
      "import.on_conflict": "On conflict",
      "import.skip": "Skip duplicates",
      "import.update": "Update existing",
      "import.error": "Report error",
      "import.upload": "Upload & Validate",
      "import.execute": "Import",
      "import.cancel": "Cancel",
      "import.valid": "{count} records valid — ready to import",
      "import.errors": "{count} errors:",
      "import.fix_errors": "Fix errors first",
      "confirm.delete": "Delete {count} selected record(s)?",
      "confirm.export": "Export {count} selected record(s)?",
      "toast.select_one": "Select at least one record",
      "toast.export_started": "Export started",
      "toast.import_success": "Imported: {imported}, Skipped: {skipped}, Updated: {updated}",
      "toast.no_file": "Select a file",
      "search.placeholder": "Search records",
      "status.selected": "{count} selected",
      "status.page": "Page {current} of {total}",
      "status.live_data": "Live data"
    },
    es: {
      "nav.overview": "Panel Principal",
      "nav.data_studio": "Estudio de Datos",
      "nav.system_pulse": "Pulso del Sistema",
      "nav.network_inspector": "Inspector de Red",
      "nav.infra_manager": "Gestor de Infra",
      "nav.health": "Salud",
      "nav.access_control": "Control de Acceso",
      "nav.audit_log": "Registro de Auditoría",
      "nav.migrations": "Migraciones",
      "nav.deployment": "Despliegue",
      "nav.jobs": "Tareas",
      "nav.cache": "Caché",
      "nav.storage": "Almacenamiento",
      "nav.sites": "Sitios",
      "cmd.search": "Buscar modelos, motores o acciones",
      "theme.dark": "Oscuro",
      "theme.light": "Claro",
      "btn.refresh": "Actualizar",
      "btn.new_record": "Nuevo registro",
      "btn.export_selected": "Exportar seleccionados",
      "btn.export_all": "Exportar",
      "btn.import": "Importar",
      "btn.delete_selected": "Eliminar seleccionados",
      "btn.csv": "CSV",
      "runtime.all_systems": "Todo nominal",
      "export.title": "Exportar Datos",
      "export.format": "Formato",
      "export.now": "Exportar Ahora",
      "export.cancel": "Cancelar",
      "export.complete": "Exportado: {records} registros",
      "import.title": "Importar Datos",
      "import.file": "Archivo (CSV o JSON)",
      "import.on_conflict": "En conflicto",
      "import.skip": "Omitir duplicados",
      "import.update": "Actualizar existentes",
      "import.error": "Reportar error",
      "import.upload": "Subir y Validar",
      "import.execute": "Importar",
      "import.cancel": "Cancelar",
      "import.valid": "{count} registros válidos — listo para importar",
      "import.errors": "{count} errores:",
      "import.fix_errors": "Corrige errores primero",
      "confirm.delete": "¿Eliminar {count} registro(s)?",
      "confirm.export": "¿Exportar {count} registro(s)?",
      "toast.select_one": "Selecciona al menos un registro",
      "toast.export_started": "Exportación iniciada",
      "toast.import_success": "Importados: {imported}, Omitidos: {skipped}, Actualizados: {updated}",
      "toast.no_file": "Selecciona un archivo",
      "search.placeholder": "Buscar registros",
      "status.selected": "{count} seleccionados",
      "status.page": "Página {current} de {total}",
      "status.live_data": "Datos en vivo"
    }
  };

  var currentLocale = "en";

  // Detect locale from localStorage or browser
  function detectLocale() {
    var stored = localStorage.getItem("gf-admin-locale");
    if (stored && translations[stored]) {
      return stored;
    }
    var browser = (navigator.language || navigator.userLanguage || "en").substring(0, 2);
    if (translations[browser]) {
      return browser;
    }
    return "en";
  }

  currentLocale = detectLocale();

  window.AdminI18n = {
    get: function(key) {
      var dict = translations[currentLocale] || translations.en;
      return dict[key] || translations.en[key] || key;
    },
    t: function(key, params) {
      var str = window.AdminI18n.get(key);
      if (!params) return str;
      return str.replace(/\{(\w+)\}/g, function(match, p) {
        return params[p] !== undefined ? params[p] : match;
      });
    },
    setLocale: function(locale) {
      if (translations[locale]) {
        currentLocale = locale;
        localStorage.setItem("gf-admin-locale", locale);
      }
    },
    getLocale: function() {
      return currentLocale;
    },
    getAvailableLocales: function() {
      return Object.keys(translations);
    }
  };
})();
