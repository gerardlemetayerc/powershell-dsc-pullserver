function renderNodeActionBlock() {
    const nodeId = getNodeIdFromUrl();
    if(!nodeId) return '';
    return '<button id="delete-node-btn" class="btn btn-danger btn-sm">Supprimer ce noeud</button>';
}

function showNodeMessage(msg, type) {
    // type: 'success' | 'error' | 'info'
    let color = type === 'success' ? '#d4edda' : (type === 'error' ? '#f8d7da' : '#e2e3e5');
    let txtColor = type === 'success' ? '#155724' : (type === 'error' ? '#721c24' : '#383d41');
    $('#node-message-block').html('<div style="background:'+color+';color:'+txtColor+';padding:8px 12px;border-radius:4px;margin-bottom:8px;">'+msg+'</div>');
}

$(document).on('click', '#delete-node-btn', function() {
    const nodeId = getNodeIdFromUrl();
    if(!nodeId) return;
    if(confirm('Êtes-vous sûr de vouloir supprimer ce noeud ? Cette action supprimera aussi ses tags et historiques de rapport.')) {
        $.ajax({
            url: '/api/v1/agents/' + encodeURIComponent(nodeId),
            type: 'DELETE',
            success: function() {
                showNodeMessage('Noeud supprimé.', 'success');
                setTimeout(function(){ window.location.href = '/web'; }, 1200);
            },
            error: function(xhr) {
                showNodeMessage('Erreur lors de la suppression du noeud: ' + (xhr.responseText || xhr.status), 'error');
            }
        });
    }
});
// node.js (frontend) pour page /web/node/{id}

function getNodeIdFromUrl() {
    // Extrait l'ID de l'URL /web/node/{id}
    const match = window.location.pathname.match(/\/node\/([^\/]+)/);
    return match ? match[1] : null;
}


// --- React-like state and render functions ---
const state = {
    agent: null,
    lastReport: null,
    reports: [],
    error: null,
    tags: {},
    configBindings: [],
    availableConfigurations: [],
};

function escapeHtml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}
function renderAgentTags() {
    let html = '';
    if (!state.tags || Object.keys(state.tags).length === 0) {
        html = '<em>Aucun tag</em>';
    } else {
        html = '<ul class="list-group mb-2" style="max-width:480px;">';
        for (const [k, values] of Object.entries(state.tags)) {
            if (Array.isArray(values)) {
                for (const v of values) {
                    html += `<li class="list-group-item py-1 px-2 d-flex justify-content-between align-items-center" style="font-size:0.97em;">
                        <span><span class="badge badge-secondary mr-2" style="font-size:0.93em;">${k}</span> <span>${v}</span></span>
                        <button class="btn btn-outline-danger btn-xs delete-tag-btn" data-key="${k}" data-value="${v}" style="padding:2px 8px;font-size:0.92em;">✕</button>
                    </li>`;
                }
            } else {
                html += `<li class="list-group-item py-1 px-2 d-flex justify-content-between align-items-center" style="font-size:0.97em;">
                    <span><span class="badge badge-secondary mr-2" style="font-size:0.93em;">${k}</span> <span>${values}</span></span>
                    <button class="btn btn-outline-danger btn-xs delete-tag-btn" data-key="${k}" data-value="${values}" style="padding:2px 8px;font-size:0.92em;">✕</button>
                </li>`;
            }
        }
        html += '</ul>';
    }
    return html;
}

function fillAgentInfo() {
    if (!state.agent) {
        $('#agent-name').text('Error loading');
        $('#agent-thumbprint').text('');
        $('#agent-lastcomm').text('');
        $('#agent-config-block').html('');
        return;
    }
    let certExpRaw = state.agent.certificate_notafter || '';
    let certExp = certExpRaw;
    let daysLeft = '';
    if (certExpRaw) {
        let expDate = new Date(certExpRaw);
        if (!isNaN(expDate.getTime())) {
            let y = expDate.getFullYear();
            let m = (expDate.getMonth()+1).toString().padStart(2,'0');
            let d = expDate.getDate().toString().padStart(2,'0');
            let h = expDate.getHours().toString().padStart(2,'0');
            let min = expDate.getMinutes().toString().padStart(2,'0');
            let s = expDate.getSeconds().toString().padStart(2,'0');
            certExp = `${y}-${m}-${d} ${h}:${min}:${s}`;
            let now = new Date();
            let diffMs = expDate - now;
            let diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));
            daysLeft = ` (${diffDays} day${diffDays === 1 ? '' : 's'} remaining)`;
        }
    }
    let lastCommRaw = state.agent.last_communication || '';
    let lastComm = lastCommRaw;
    if (lastCommRaw) {
        let dt = new Date(lastCommRaw);
        if (!isNaN(dt.getTime())) {
            let y = dt.getFullYear();
            let m = (dt.getMonth()+1).toString().padStart(2,'0');
            let d = dt.getDate().toString().padStart(2,'0');
            let h = dt.getHours().toString().padStart(2,'0');
            let min = dt.getMinutes().toString().padStart(2,'0');
            let s = dt.getSeconds().toString().padStart(2,'0');
            lastComm = `${y}-${m}-${d} ${h}:${min}:${s}`;
        }
    }
    $('#agent-name').text(state.agent.node_name || '');
    $('#agent-thumbprint').text(certExp + daysLeft);
    $('#agent-lastcomm').text(lastComm);
    if (Array.isArray(state.agent.configuration_bindings) && state.agent.configuration_bindings.length > 0) {
        let html = `<strong>Configurations:</strong> ` +
            state.agent.configuration_bindings.map(c => {
                const badge = c.schedule_type && c.schedule_type !== 'none' ? 'badge-warning' : 'badge-info';
                const suffix = c.schedule_type && c.schedule_type !== 'none' ? ` (${c.schedule_type})` : ' (main)';
                return `<span class="badge ${badge} mr-1">${c.configuration_name}${suffix}</span>`;
            }).join(' ');
        $('#agent-config-block').html(html);
    } else if (Array.isArray(state.agent.configurations) && state.agent.configurations.length > 0) {
        let html = `<strong>Configuration${state.agent.configurations.length > 1 ? 's' : ''} associée${state.agent.configurations.length > 1 ? 's' : ''} :</strong> ` +
            state.agent.configurations.map(c => `<span class="badge badge-info mr-1">${c}</span>`).join(' ');
        $('#agent-config-block').html(html);
    } else {
        $('#agent-config-block').html('');
    }
}

function formatScheduleDate(value) {
    if (!value) return '-';
    const normalized = value.replace(' ', 'T');
    const dt = new Date(normalized.endsWith('Z') ? normalized : normalized + 'Z');
    if (Number.isNaN(dt.getTime())) return value;
    const y = dt.getFullYear();
    const m = String(dt.getMonth() + 1).padStart(2, '0');
    const d = String(dt.getDate()).padStart(2, '0');
    const hh = String(dt.getHours()).padStart(2, '0');
    const mm = String(dt.getMinutes()).padStart(2, '0');
    const tzPart = Intl.DateTimeFormat(undefined, { timeZoneName: 'short' })
        .formatToParts(dt)
        .find((part) => part.type === 'timeZoneName');
    const tz = tzPart ? tzPart.value : 'local';
    return `${y}-${m}-${d} ${hh}:${mm} ${tz}`;
}

function formatNextExecutionWindow(binding) {
    if (!binding || !binding.enabled) {
        return '-';
    }
    if (binding.next_execution_completed) {
        return 'Completed';
    }
    if (!binding.next_execution_start_at || !binding.next_execution_end_at) {
        return '-';
    }
    return `${formatScheduleDate(binding.next_execution_start_at)} -> ${formatScheduleDate(binding.next_execution_end_at)}`;
}

function renderExecutionStatusBadge(binding) {
    const rawStatus = (binding && binding.last_execution_status) ? String(binding.last_execution_status) : '';
    if (!rawStatus) {
        return '<span class="badge badge-secondary">No run yet</span>';
    }

    const normalized = rawStatus.toLowerCase();
    let badgeClass = 'badge-secondary';
    if (normalized === 'success' || normalized === 'ok') {
        badgeClass = 'badge-success';
    } else if (normalized === 'failure' || normalized === 'failed' || normalized === 'error') {
        badgeClass = 'badge-danger';
    } else if (normalized === 'pending_apply' || normalized === 'getconfiguration') {
        badgeClass = 'badge-warning';
    }

    return `<span class="badge ${badgeClass}">${rawStatus}</span>`;
}

function renderConfigBindingsTable() {
    if (!Array.isArray(state.configBindings) || state.configBindings.length === 0) {
        return '<em>No configuration binding configured.</em>';
    }

    let html = '<table class="table table-sm table-bordered"><thead><tr>' +
        '<th>Configuration</th><th>Type</th><th>Scheduled at</th><th>Next execution window</th><th>Last execution</th><th>Executed at</th><th>Recurrence (min)</th><th>Window (min)</th><th>Enabled</th><th>Actions</th>' +
        '</tr></thead><tbody>';

    state.configBindings.forEach((b) => {
        html += `<tr>
            <td>${b.configuration_name || ''}</td>
            <td>${b.schedule_type || 'none'}</td>
            <td>${formatScheduleDate(b.scheduled_at)}</td>
            <td>${formatNextExecutionWindow(b)}</td>
            <td>${renderExecutionStatusBadge(b)}</td>
            <td>${formatScheduleDate(b.last_execution_at)}</td>
            <td>${b.recurrence_minutes || '-'}</td>
            <td>${b.window_minutes || 30}</td>
            <td>${b.enabled ? 'Yes' : 'No'}</td>
            <td>
                <button class="btn btn-outline-primary btn-xs edit-binding-btn" data-name="${(b.configuration_name || '').replace(/"/g, '&quot;')}">Edit</button>
                <button class="btn btn-outline-danger btn-xs delete-binding-btn" data-name="${(b.configuration_name || '').replace(/"/g, '&quot;')}">Delete</button>
            </td>
        </tr>`;
    });

    html += '</tbody></table>';
    return html;
}

function renderConfigBindings() {
    $('#config-bindings-block').html(renderConfigBindingsTable());
}

function initializeConfigurationSelect2() {
    const $select = $('#config-binding-name');
    if (!$select.length || typeof $.fn.select2 !== 'function') {
        return;
    }
    if ($select.data('select2')) {
        return;
    }
    $select.select2({
        theme: 'bootstrap4',
        width: '100%',
        placeholder: 'Sélectionner une configuration',
        allowClear: true
    });
}

function renderConfigurationNameOptions() {
    const options = Array.isArray(state.availableConfigurations) ? state.availableConfigurations : [];
    const currentSelection = $('#config-binding-name').val();
    const html = ['<option value="">Sélectionner une configuration</option>']
        .concat(options.map((name) => `<option value="${escapeHtml(name)}">${escapeHtml(name)}</option>`))
        .join('');
    $('#config-binding-name').html(html);

    if (currentSelection && options.includes(currentSelection)) {
        $('#config-binding-name').val(currentSelection);
    } else {
        $('#config-binding-name').val('');
    }

    initializeConfigurationSelect2();
    if ($('#config-binding-name').data('select2')) {
        $('#config-binding-name').trigger('change.select2');
    }
}

function showConfigBindingMessage(msg, type) {
    const cls = type === 'error' ? 'text-danger' : 'text-success';
    $('#config-binding-message').removeClass('text-danger text-success').addClass(cls).text(msg || '');
}

function getConfigBindingByName(name) {
    if (!name) return null;
    return (state.configBindings || []).find((binding) => binding.configuration_name === name) || null;
}

function hasMainBinding(excludeName) {
    const excluded = (excludeName || '').toLowerCase();
    return (state.configBindings || []).some((binding) => {
        const scheduleType = (binding.schedule_type || 'none').toLowerCase();
        const bindingName = (binding.configuration_name || '').toLowerCase();
        if (scheduleType !== 'none') {
            return false;
        }
        if (excluded && bindingName === excluded) {
            return false;
        }
        return true;
    });
}

function isMainTypeAllowedForCurrentForm() {
    return true;
}

function refreshConfigBindingTypeOptions() {
    const $typeSelect = $('#config-binding-type');
    const $mainOption = $typeSelect.find('option[value="none"]');
    if (!$typeSelect.length || !$mainOption.length) {
        return;
    }

    $mainOption.prop('disabled', false);
    $mainOption.text('Main');
}

function resetConfigBindingForm() {
    $('#config-binding-editing-original').val('');
    renderConfigurationNameOptions();
    $('#config-binding-name').val('');
    if ($('#config-binding-name').data('select2')) {
        $('#config-binding-name').trigger('change.select2');
    }
    $('#config-binding-type').val('none');
    refreshConfigBindingTypeOptions();
    $('#config-binding-scheduled-at').val('');
    $('#config-binding-recurrence').val('');
    $('#config-binding-window').val('30');
    $('#config-binding-enabled').prop('checked', true);
    showConfigBindingMessage('', 'success');
    updateConfigBindingFormByType();
}

function updateConfigBindingFormByType() {
    const t = ($('#config-binding-type').val() || 'none').toLowerCase();
    const isScheduled = t === 'oneshot' || t === 'recurring';
    const isRecurring = t === 'recurring';
    $('#config-binding-scheduled-at').prop('disabled', !isScheduled);
    $('#config-binding-recurrence').prop('disabled', !isRecurring);
}

function toInputDateTime(value) {
    if (!value) return '';
    const normalized = value.replace(' ', 'T');
    const dt = new Date(normalized.endsWith('Z') ? normalized : normalized + 'Z');
    if (Number.isNaN(dt.getTime())) return '';
    const y = dt.getFullYear();
    const m = String(dt.getMonth() + 1).padStart(2, '0');
    const d = String(dt.getDate()).padStart(2, '0');
    const hh = String(dt.getHours()).padStart(2, '0');
    const mm = String(dt.getMinutes()).padStart(2, '0');
    return `${y}-${m}-${d}T${hh}:${mm}`;
}

function loadConfigBindings(agentId) {
    return $.ajax({
        url: `/api/v1/agents/${agentId}/configs`,
        method: 'GET',
        dataType: 'json',
        success: function(bindings) {
            state.configBindings = Array.isArray(bindings) ? bindings : [];
            renderConfigBindings();
            refreshConfigBindingTypeOptions();
        },
        error: function() {
            state.configBindings = [];
            renderConfigBindings();
            refreshConfigBindingTypeOptions();
        }
    });
}

function loadAvailableConfigurations() {
    return $.ajax({
        url: '/api/v1/configuration_models?current=1',
        method: 'GET',
        dataType: 'json',
        success: function(configurations) {
            const names = Array.isArray(configurations)
                ? configurations.map((configuration) => configuration && configuration.name).filter(Boolean)
                : [];
            state.availableConfigurations = [...new Set(names)].sort((left, right) => left.localeCompare(right));
            renderConfigurationNameOptions();
        },
        error: function() {
            state.availableConfigurations = [];
            renderConfigurationNameOptions();
        }
    });
}

function renderSelectedReport() {
    const report = state.selectedReport;
    if (!report) {
        return `<p><strong>Status:</strong> <span id="report-status">No report</span></p>
                <div><strong>Start Time:</strong> <span id="report-starttime"></span></div>
                <div><strong>Operation Type:</strong> <span id="report-operationtype"></span></div>
                <p><strong>Errors:</strong> <span id="report-errors"></span></p>
                <pre id="report-json" style="background:#eee; padding:10px;"></pre>`;
    }
    // Extraction des ressources en/desired state depuis StatusData
    let resourcesInDesired = [];
    let resourcesNotInDesired = [];
    if (Array.isArray(report.StatusData) && report.StatusData.length > 0) {
        for (const sd of report.StatusData) {
            try {
                const obj = typeof sd === 'string' ? JSON.parse(sd) : sd;
                if (Array.isArray(obj.ResourcesInDesiredState)) {
                    resourcesInDesired = resourcesInDesired.concat(obj.ResourcesInDesiredState);
                }
                if (Array.isArray(obj.ResourcesNotInDesiredState)) {
                    resourcesNotInDesired = resourcesNotInDesired.concat(obj.ResourcesNotInDesiredState);
                }
            } catch(e) {}
        }
    }
    function renderResourceTable(resources, title) {
        if (!Array.isArray(resources) || resources.length === 0) return '';
        let cols = Object.keys(resources[0] || {});
        // Retire la colonne StartDate, ModuleVersion, ResourceId et SourceInfo si présentes
        cols = cols.filter(c => c !== 'StartDate' && c !== 'ModuleVersion' && c !== 'ResourceId' && c !== 'SourceInfo');
        let html = `<p><strong>${title}</strong></p><table style="width:100%" class="table table-bordered table-sm"><thead><tr>`;
        html += cols.map(c => `<th>${c}</th>`).join('');
        html += '</tr></thead><tbody>';
        const isNotDesired = title.toLowerCase().includes('not in desired');
        for(const [rowIdx, res] of resources.entries()) {
            html += '<tr>' + cols.map(c => {
                if (c === 'ModuleName') {
                    let mod = res[c] || '';
                    let version = res['ModuleVersion'] || '';
                    if (mod) {
                        return `<td>${mod}${version ? '<span class="badge badge-info" style="font-size:90%; margin-left:6px;">' + version + '</span>' : ''}</td>`;
                    } else {
                        return `<td></td>`;
                    }
                } else if (c.toLowerCase() === 'error' && isNotDesired) {
                    let err = res[c] || '';
                    let shortErr = err.length > 60 ? err.substring(0, 60) + '…' : err;
                    let btn = err ? `<a href="#" class="show-error-modal" data-error="${encodeURIComponent(err)}" data-row="${rowIdx}">${shortErr}</a>` : '';
                    return `<td>${btn}</td>`;
                } else {
                    return `<td>${res[c] || ''}</td>`;
                }
            }).join('') + '</tr>';
        }
        html += '</tbody></table>';
        return html;
    }
    // Statut et infos génériques
    let statusValue = report.Status || report.status || '';
    let statusText = statusValue ? statusValue : 'Unknown';
    let statusClass = '';
    if (statusValue.toLowerCase() === 'success') {
        statusClass = 'alert alert-success';
    } else if (statusValue) {
        statusClass = 'alert alert-danger';
    } else {
        statusClass = 'alert alert-secondary';
    }
    // Robust extraction of StartTime/EndTime
    let startTime = report.StartTime || report.startTime || '';
    let endTime = report.EndTime || report.endTime || '';
    if (!startTime && Array.isArray(report.StatusData) && report.StatusData.length > 0) {
        try {
            const statusObj = JSON.parse(report.StatusData[0]);
            startTime = statusObj.StartDate || '';
        } catch(e) {}
    }
    if (!startTime && typeof state.selectedReportIdx === 'number' && state.reports[state.selectedReportIdx]) {
        startTime = state.reports[state.selectedReportIdx].created_at || '';
    }
    let operationType = report.OperationType || report.operationType || '';

    // Partie détails/erreurs
    let hasErrors = Array.isArray(report.Errors) && report.Errors.length > 0;
    let errorsHtml = '';
    if (hasErrors) {
        let table = '<table class="table table-bordered table-sm"><thead><tr>';
        let cols = [];
        try {
            const first = JSON.parse(report.Errors[0]);
            cols = Object.keys(first);
            table += cols.map(c => `<th>${c}</th>`).join('');
            table += '</tr></thead><tbody>';
            for(const errStr of report.Errors) {
                try {
                    const err = JSON.parse(errStr);
                    table += '<tr>' + cols.map(c => `<td>${err[c] || ''}</td>`).join('') + '</tr>';
                } catch(e) {
                    table += `<tr><td colspan="${cols.length}">${errStr}</td></tr>`;
                }
            }
            table += '</tbody></table>';
        } catch(e) {
            table = report.Errors.join('<br>');
        }
        errorsHtml = `<p><strong>Errors:</strong></p>${table}`;
    }

    return `
        <div class="${statusClass}" style="padding:8px 12px;"><strong>Status:</strong> <span id="report-status">${statusText}</span></div>
        <div><strong>Start Time:</strong> <span id="report-starttime">${startTime}</span></div>
        <div><strong>End Time:</strong> <span id="report-endtime">${endTime}</span></div>
        <div><strong>Operation Type:</strong> <span id="report-operationtype">${operationType}</span></div>
        <hr>
        ${renderResourceTable(resourcesInDesired, 'Resources In Desired State')}
        ${renderResourceTable(resourcesNotInDesired, 'Resources Not In Desired State')}
        ${errorsHtml}
        <div style="margin-top:1em;">
            <button id="toggle-raw-json" class="btn btn-outline-secondary btn-sm" type="button" style="margin-bottom:5px;">Afficher les données brutes</button>
            <div id="raw-json-block" style="display:none; max-width:100%; overflow-x:auto;">
                <pre id="report-json" style="background:#eee; padding:10px;">${JSON.stringify(report, null, 2)}</pre>
            </div>
        </div>
                <script>
                        $(function(){
                                $('#toggle-raw-json').off('click').on('click', function() {
                                        const block = $('#raw-json-block');
                                        if(block.is(':visible')) {
                                                block.slideUp(150);
                                                $(this).text('Afficher les données brutes');
                                        } else {
                                                block.slideDown(150);
                                                $(this).text('Masquer les données brutes');
                                        }
                                });
                                // Gestion modale erreur détaillée
                                $(document).off('click', '.show-error-modal').on('click', '.show-error-modal', function(e) {
                                        e.preventDefault();
                                        var err = decodeURIComponent($(this).data('error'));
                                        if($('#errorDetailModal').length === 0) {
                                                var modalHtml = '<div class="modal fade" id="errorDetailModal" tabindex="-1" role="dialog" aria-labelledby="errorDetailModalLabel" aria-hidden="true">'+
                                                    '<div class="modal-dialog modal-lg" role="document">'+
                                                        '<div class="modal-content">'+
                                                            '<div class="modal-header bg-danger text-white">'+
                                                                '<h5 class="modal-title" id="errorDetailModalLabel">Erreur complète</h5>'+ 
                                                                '<button type="button" class="close text-white" data-dismiss="modal" aria-label="Close">'+
                                                                    '<span aria-hidden="true">&times;</span>'+ 
                                                                '</button>'+ 
                                                            '</div>'+ 
                                                            '<div class="modal-body" id="errorDetailModalBody" style="background:#222;color:#fff;white-space:pre-wrap;font-family:monospace;font-size:1em;max-height:70vh;overflow:auto;border-radius:6px;padding:18px 16px;"></div>'+ 
                                                        '</div>'+ 
                                                    '</div>'+ 
                                                '</div>';
                                                $('body').append(modalHtml);
                                        }
                                        // Formatage JSON si possible
                                        let formatted = err;
                                        try {
                                                let obj = JSON.parse(err);
                                                formatted = JSON.stringify(obj, null, 2);
                                        } catch(e) {}
                                        $('#errorDetailModalBody').text(formatted);
                                        $('#errorDetailModal').modal('show');
                                });
                        });
                </script>
    `;
}

function renderReportsDropdown() {
    if (!Array.isArray(state.reports) || state.reports.length === 0) {
        return '<em>No report available.</em>';
    }
    let html = '<label for="report-select">Select a report:</label> ';
    html += '<select id="report-select" class="form-control form-control-sm" style="width:auto; display:inline-block; margin-left:10px;">';
    state.reports.forEach((rep, idx) => {
        const label = rep.created_at || '';
        let status = (rep.Status || rep.status || '').toLowerCase();
        let bgColor = '';
        if (status === 'success') {
            bgColor = 'background-color:#d4edda; color:#155724;'; // vert pâle
        } else if (status) {
            bgColor = 'background-color:#f8d7da; color:#721c24;'; // rouge pâle
        } else {
            bgColor = 'background-color:#e9ecef; color:#495057;'; // gris clair
        }
        html += `<option value="${idx}"${state.selectedReportIdx === idx ? ' selected' : ''} style="${bgColor}">${label}</option>`;
    });
    html += '</select>';
    return html;
}

function renderAll() {
    fillAgentInfo();
    $('#agent-tags-block').html(renderAgentTags());
    renderConfigBindings();
    $('#reports-dropdown-block').html(renderReportsDropdown());
    $('#last-report-block').html(renderSelectedReport());
    $('#node-action-block').html(renderNodeActionBlock());
}

$(document).ready(function() {
    const agentId = getNodeIdFromUrl();
    // Ne pas réécrire le HTML, juste remplir les valeurs
    renderAll();
    if (!agentId) {
        state.error = 'Agent ID not found in URL';
        renderAll();
        return;
    }
    // Load agent info (utilise $.ajax pour profiter de la config globale JWT)
    $.ajax({
        url: `/api/v1/agents/${agentId}`,
        method: 'GET',
        dataType: 'json',
        success: function(agent) {
            state.agent = agent;
            state.configBindings = Array.isArray(agent.configuration_bindings) ? agent.configuration_bindings : state.configBindings;
            renderAll();
        },
        error: function() {
            state.agent = null;
            renderAll();
        }
    });
    // Load tags
    $.ajax({
        url: `/api/v1/agents/${agentId}/tags`,
        method: 'GET',
        dataType: 'json',
        success: function(tags) {
            state.tags = tags || {};
            renderAll();
        },
        error: function() {
            state.tags = {};
            renderAll();
        }
    });
    loadConfigBindings(agentId);
    loadAvailableConfigurations();
    resetConfigBindingForm();
        // Ajout d'un tag
        $(document).on('submit', '#add-tag-form', function(e) {
            e.preventDefault();
            const key = $('#tag-key').val().trim();
            const value = $('#tag-value').val().trim();
            if (!key || !value) return;
            $.ajax({
                url: `/api/v1/agents/${agentId}/tags`,
                method: 'PUT',
                contentType: 'application/json',
                data: JSON.stringify({ key, value }),
                success: function() {
                    $('#tag-key').val('');
                    $('#tag-value').val('');
                    // Reload tags
                    $.ajax({
                        url: `/api/v1/agents/${agentId}/tags`,
                        method: 'GET',
                        dataType: 'json',
                        success: function(tags) {
                            state.tags = tags || {};
                            renderAll();
                        }
                    });
                }
            });
        });
        // Suppression d'une valeur de tag
        $(document).on('click', '.delete-tag-btn', function() {
            const key = $(this).data('key');
            const value = $(this).data('value');
            if (!key || !value) return;
            $.ajax({
                url: `/api/v1/agents/${agentId}/tags?key=` + encodeURIComponent(key) + `&value=` + encodeURIComponent(value),
                method: 'DELETE',
                success: function() {
                    // Reload tags
                    $.ajax({
                        url: `/api/v1/agents/${agentId}/tags`,
                        method: 'GET',
                        dataType: 'json',
                        success: function(tags) {
                            state.tags = tags || {};
                            renderAll();
                        }
                    });
                }
            });
        });
    // Charger la liste des rapports
    $.ajax({
        url: `/api/v1/agents/${agentId}/reports?mofapplied=1`,
        method: 'GET',
        dataType: 'json',
        success: function(reports) {
            state.reports = reports;
            state.selectedReportIdx = 0;
            renderAll();
            setTimeout(() => {
                $('#report-select').trigger('change');
            }, 0);
        },
        error: function() {
            state.reports = [];
            state.selectedReportIdx = null;
            state.selectedReport = null;
            renderAll();
        }
    });

    // Gestion du changement de sélection de rapport (delegation)
    $(document).on('change', '#report-select', function() {
        const idx = parseInt($(this).val(), 10);
        state.selectedReportIdx = idx;
        const rep = state.reports[idx];
        if (rep && rep.job_id) {
            // Aller chercher le rapport détaillé
            $.ajax({
                url: `/api/v1/agents/${agentId}/reports/${rep.job_id}`,
                method: 'GET',
                dataType: 'json',
                success: function(report) {
                    state.selectedReport = report;
                    renderAll();
                },
                error: function() {
                    state.selectedReport = null;
                    renderAll();
                }
            });
        } else {
            state.selectedReport = rep || null;
            renderAll();
        }
    });

       // Bouton gestion propriétés du nœud
       $(document).on('click', '#node-properties-link', function(e) {
           e.preventDefault();
           if(agentId) {
               window.location.href = '/web/node/' + encodeURIComponent(agentId) + '/properties';
           }
       });

    $(document).on('change', '#config-binding-type', function() {
        updateConfigBindingFormByType();
    });

    $(document).on('click', '#config-binding-reset', function() {
        resetConfigBindingForm();
    });

    $(document).on('click', '.edit-binding-btn', function() {
        const name = $(this).data('name');
        const binding = (state.configBindings || []).find((b) => b.configuration_name === name);
        if (!binding) return;
        $('#config-binding-editing-original').val(binding.configuration_name || '');
        renderConfigurationNameOptions();
        $('#config-binding-name').val(binding.configuration_name || '');
        if ($('#config-binding-name').data('select2')) {
            $('#config-binding-name').trigger('change.select2');
        }
        $('#config-binding-type').val(binding.schedule_type || 'none');
        refreshConfigBindingTypeOptions();
        $('#config-binding-scheduled-at').val(toInputDateTime(binding.scheduled_at));
        $('#config-binding-recurrence').val(binding.recurrence_minutes || '');
        $('#config-binding-window').val(binding.window_minutes || 30);
        $('#config-binding-enabled').prop('checked', !!binding.enabled);
        updateConfigBindingFormByType();
        showConfigBindingMessage('', 'success');
    });

    $(document).on('click', '.delete-binding-btn', function() {
        const name = $(this).data('name');
        if (!name) return;
        if (!confirm(`Supprimer la configuration "${name}" ?`)) return;
        $.ajax({
            url: `/api/v1/agents/${agentId}/configs`,
            method: 'DELETE',
            contentType: 'application/json',
            data: JSON.stringify({ configuration_name: name }),
            success: function() {
                showConfigBindingMessage('Configuration supprimée.', 'success');
                loadConfigBindings(agentId);
                $.ajax({
                    url: `/api/v1/agents/${agentId}`,
                    method: 'GET',
                    dataType: 'json',
                    success: function(agent) {
                        state.agent = agent;
                        renderAll();
                    }
                });
            },
            error: function(xhr) {
                showConfigBindingMessage(`Erreur suppression: ${xhr.responseText || xhr.status}`, 'error');
            }
        });
    });

    $(document).on('submit', '#config-binding-form', function(e) {
        e.preventDefault();
        const configurationName = ($('#config-binding-name').val() || '').trim();
        const scheduleType = ($('#config-binding-type').val() || 'none').toLowerCase();
        const scheduledAt = ($('#config-binding-scheduled-at').val() || '').trim();
        const recurrenceMinutesRaw = ($('#config-binding-recurrence').val() || '').trim();
        const windowMinutesRaw = ($('#config-binding-window').val() || '').trim();
        const enabled = $('#config-binding-enabled').is(':checked');
        const originalName = ($('#config-binding-editing-original').val() || '').trim();
        const availableConfigurations = new Set(state.availableConfigurations || []);

        if (!configurationName) {
            showConfigBindingMessage('Le nom de configuration est requis.', 'error');
            return;
        }
        if (!availableConfigurations.has(configurationName)) {
            showConfigBindingMessage('Veuillez sélectionner une configuration existante.', 'error');
            return;
        }
        if ((scheduleType === 'oneshot' || scheduleType === 'recurring') && !scheduledAt) {
            showConfigBindingMessage('La date planifiée est requise.', 'error');
            return;
        }
        if (scheduleType === 'recurring' && (!recurrenceMinutesRaw || Number(recurrenceMinutesRaw) <= 0)) {
            showConfigBindingMessage('La récurrence doit être > 0.', 'error');
            return;
        }

        const payload = {
            configuration_name: configurationName,
            schedule_type: scheduleType,
            window_minutes: windowMinutesRaw ? Number(windowMinutesRaw) : 30,
            enabled: enabled
        };
        if (scheduledAt) {
            const dt = new Date(scheduledAt);
            if (!Number.isNaN(dt.getTime())) {
                payload.scheduled_at = dt.toISOString();
            } else {
                payload.scheduled_at = scheduledAt;
            }
        }
        if (recurrenceMinutesRaw) {
            payload.recurrence_minutes = Number(recurrenceMinutesRaw);
        }

        const saveBinding = function() {
            $.ajax({
                url: `/api/v1/agents/${agentId}/configs`,
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify(payload),
                success: function() {
                    showConfigBindingMessage('Configuration sauvegardée.', 'success');
                    resetConfigBindingForm();
                    loadConfigBindings(agentId);
                    $.ajax({
                        url: `/api/v1/agents/${agentId}`,
                        method: 'GET',
                        dataType: 'json',
                        success: function(agent) {
                            state.agent = agent;
                            renderAll();
                        }
                    });
                },
                error: function(xhr) {
                    showConfigBindingMessage(`Erreur sauvegarde: ${xhr.responseText || xhr.status}`, 'error');
                }
            });
        };

        if (originalName && originalName !== configurationName) {
            $.ajax({
                url: `/api/v1/agents/${agentId}/configs`,
                method: 'DELETE',
                contentType: 'application/json',
                data: JSON.stringify({ configuration_name: originalName }),
                complete: saveBinding
            });
            return;
        }

        saveBinding();
    });
});
