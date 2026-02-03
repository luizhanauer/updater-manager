<script setup>
import { onMounted, ref, computed } from 'vue';
import { GetApps, ToggleInstall, ToggleAutoUpdate, RefreshCatalog } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const apps = ref([]);
const lastCheck = ref("");
const isRefreshing = ref(false);
const fallbackIcon = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/d9/Icon-round-Question_mark.svg/1024px-Icon-round-Question_mark.svg.png";

const refresh = async () => {
  try { apps.value = await GetApps() } 
  catch (e) { console.error(e) }
};

const manualRefresh = async () => {
  isRefreshing.value = true;
  await RefreshCatalog();
  setTimeout(() => isRefreshing.value = false, 2000);
};

const install = async (app) => {
  if (app.status === 'downloading' || app.status === 'installing') return;
  app.status = 'waiting'; 
  app.message = 'Aguardando...';
  await ToggleInstall(app.id, !app.is_installed);
};

const toggleAuto = async (app) => {
  await ToggleAutoUpdate(app.id, app.auto_update);
};

const onImgError = (event) => { event.target.src = fallbackIcon; };

const formattedDate = computed(() => {
  if (!lastCheck.value || lastCheck.value === "0001-01-01T00:00:00Z") return "Nunca";
  const date = new Date(lastCheck.value);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
});

// Helper de formatação JS (para updates dinâmicos)
const fmtBytes = (bytes) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

onMounted(() => {
  refresh();
  
  EventsOn("daemon_event", (evt) => {
    if (evt.type === "status_update" || evt.type === "init") {
       
       if (evt.payload.catalog_checked_at) {
         lastCheck.value = evt.payload.catalog_checked_at;
       }

       const statusMap = evt.payload.apps;
       
       apps.value = apps.value.map(app => {
         if (statusMap[app.id]) {
           const s = statusMap[app.id];
           
           app.status = s.state;
           app.message = s.message;
           app.local_version = s.current_version;
           app.is_installed = !!s.current_version;
           app.auto_update = s.auto_update;
           
           // LÓGICA DE DISPLAY DE TAMANHO/PROGRESSO
           if (s.total_size > 0 && (s.state === 'downloading' || s.state === 'installing')) {
             // Dinâmico (SSE)
             app.progress = s.progress;
             app.size_str = `${fmtBytes(s.downloaded_size)} / ${fmtBytes(s.total_size)}`;
             app.total_size = fmtBytes(s.total_size); 

           } else if (app.total_size) {
             // Estático (Catálogo)
             // app.total_size já veio preenchido pelo GetApps
             app.size_str = app.total_size;
             app.progress = 0;
           }
         }
         return app;
       });
    }
  });
});
</script>

<template>
  <div class="main-wrapper">
    <header class="app-header">
      <div class="header-content">
        <div class="header-left">
           <h1>Central de Apps</h1>
           <div class="last-check">
             Atualizado às {{ formattedDate }}
           </div>
        </div>
        
        <div class="header-right">
           <button class="btn-refresh" @click="manualRefresh" :disabled="isRefreshing">
             <span v-if="isRefreshing" class="spin">↻</span>
             <span v-else>Verificar Atualizações</span>
           </button>
        </div>
      </div>
    </header>

    <div class="container">
      <div class="app-list">
        <div v-for="app in apps" :key="app.id" class="app-row">
          
          <div class="col-icon">
            <img :src="app.icon_url || fallbackIcon" @error="onImgError" />
          </div>
          
          <div class="col-info">
            <div class="info-top">
              <h3>{{ app.name }}</h3>
              <span class="badge-cat">{{ app.category }}</span>
            </div>
            <p class="desc">{{ app.description }}</p>
            
            <div v-if="app.status === 'downloading' || app.status === 'waiting'" class="progress-wrap">
              <div class="progress-bar" :style="{width: (app.status==='waiting' ? 100 : app.progress) + '%', opacity: app.status==='waiting' ? 0.3 : 1}"></div>
              <span class="progress-txt" v-if="app.status === 'waiting'">Solicitando...</span>
              <span class="progress-txt" v-else>{{ app.size_str }} ({{ app.progress }}%)</span>
            </div>
            
            <div v-else-if="app.message" class="status-msg" :class="app.status">
              {{ app.message }}
            </div>
          </div>

          <div class="col-meta">
            <div class="ver-line" title="Versão Instalada">
              <span class="lbl">Local:</span>
              <span class="val" :class="app.is_installed ? 'inst' : 'none'">
                {{ app.is_installed ? 'v' + app.local_version : '-' }}
              </span>
            </div>
            <div class="ver-line" title="Última Versão Disponível">
              <span class="lbl">Remoto:</span>
              <span class="val rem">v{{ app.remote_version }}</span>
            </div>
             <div class="ver-line" v-if="!app.is_installed || (app.local_version !== app.remote_version)">
              <span class="lbl">Tam:</span>
              <span class="val size">{{ app.total_size }}</span>
            </div>
          </div>

          <div class="col-actions">
            <label v-if="app.is_installed" class="toggle-switch" title="Manter Atualizado">
              <input type="checkbox" v-model="app.auto_update" @change="toggleAuto(app)">
              <span class="slider"></span>
            </label>

            <button 
              @click="install(app)" 
              :disabled="app.status === 'downloading' || app.status === 'installing' || app.status === 'waiting' || app.status === 'removing'"
              :class="['action-btn', app.is_installed ? 'btn-danger' : 'btn-primary']"
            >
              <span v-if="app.status === 'downloading'">Baixando</span>
              <span v-else-if="app.status === 'installing'">Instalando</span>
              <span v-else-if="app.status === 'removing'">Removendo</span>
              <span v-else-if="app.status === 'waiting'">...</span>
              <span v-else>{{ app.is_installed ? 'Remover' : 'Instalar' }}</span>
            </button>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>

<style>
/* Reset Global */
body { margin: 0; font-family: 'Inter', -apple-system, sans-serif; background: #f0f2f5; color: #333; }
* { box-sizing: border-box; }

/* Header */
.app-header { background: #2c3e50; color: white; padding: 15px 0; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 20px; }
.header-content { max-width: 900px; margin: 0 auto; padding: 0 20px; display: flex; justify-content: space-between; align-items: center; }
.header-left { display: flex; flex-direction: column; }
.app-header h1 { margin: 0; font-size: 1.2rem; font-weight: 600; }
.last-check { font-size: 0.75rem; opacity: 0.8; margin-top: 4px; }

.btn-refresh {
  background: rgba(255,255,255,0.15); border: 1px solid rgba(255,255,255,0.3); color: white;
  padding: 8px 16px; border-radius: 6px; cursor: pointer; font-size: 0.85rem; font-weight: 500; transition: 0.2s;
  display: flex; align-items: center; gap: 5px;
}
.btn-refresh:hover { background: rgba(255,255,255,0.25); }
.btn-refresh:disabled { opacity: 0.6; cursor: wait; }
.spin { display: inline-block; animation: spin 1s linear infinite; font-size: 1.1rem; }
@keyframes spin { 100% { transform: rotate(360deg); } }

.container { max-width: 900px; margin: 0 auto; padding: 0 20px 40px 20px; }

/* Lista */
.app-list { display: flex; flex-direction: column; gap: 10px; }

.app-row {
  display: flex; align-items: center; background: white; padding: 12px 16px; border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid #e1e4e8;
}

/* Coluna Ícone */
.col-icon { flex: 0 0 48px; margin-right: 16px; }
.col-icon img { width: 48px; height: 48px; object-fit: contain; }

/* Coluna Info */
.col-info { flex: 1; min-width: 0; margin-right: 15px; }
.info-top { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
h3 { margin: 0; font-size: 1rem; font-weight: 600; color: #2c3e50; }
.badge-cat { font-size: 0.65rem; background: #eee; color: #666; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; letter-spacing: 0.5px; }
.desc { margin: 0; font-size: 0.85rem; color: #7f8c8d; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* Status e Progresso */
.progress-wrap { background: #ecf0f1; height: 16px; border-radius: 8px; position: relative; overflow: hidden; margin-top: 5px; }
.progress-bar { background: #3498db; height: 100%; transition: width 0.3s ease; }
.progress-txt { position: absolute; width: 100%; text-align: center; top: 0; line-height: 16px; font-size: 0.7rem; font-weight: bold; color: #2c3e50; }
.status-msg { font-size: 0.8rem; margin-top: 4px; color: #7f8c8d; }
.status-msg.error { color: #c0392b; font-weight: bold; }

/* Coluna Metadados (Versões) */
.col-meta { flex: 0 0 160px; font-size: 0.8rem; border-left: 1px solid #eee; padding-left: 15px; display: flex; flex-direction: column; justify-content: center; gap: 2px; }
.ver-line { display: flex; justify-content: space-between; }
.lbl { color: #95a5a6; }
.val { font-family: 'Consolas', monospace; font-weight: 600; max-width: 100px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; text-align: right; }
.val.inst { color: #27ae60; }
.val.rem { color: #2980b9; }
.val.none { color: #bdc3c7; }
.val.size { color: #7f8c8d; }

/* Coluna Ações */
.col-actions { flex: 0 0 120px; display: flex; flex-direction: column; align-items: flex-end; gap: 8px; margin-left: 15px; }

.action-btn {
  width: 100%; padding: 6px 0; border: none; border-radius: 5px; font-size: 0.85rem; font-weight: 600; cursor: pointer; transition: 0.2s;
}
.btn-primary { background: #3498db; color: white; }
.btn-primary:hover { background: #2980b9; }
.btn-danger { background: white; border: 1px solid #e74c3c; color: #e74c3c; }
.btn-danger:hover { background: #e74c3c; color: white; }
button:disabled { opacity: 0.6; cursor: not-allowed; background: #bdc3c7; color: white; border: none; }

/* Toggle */
.toggle-switch { display: inline-block; width: 34px; height: 18px; position: relative; cursor: pointer; }
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; transition: .4s; border-radius: 18px; }
.slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 2px; bottom: 2px; background-color: white; transition: .4s; border-radius: 50%; }
input:checked + .slider { background-color: #27ae60; }
input:checked + .slider:before { transform: translateX(16px); }
</style>