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
    error: null
};

function renderAgentInfo() {
    if (!state.agent) {
        return `<p><strong>Name:</strong> <span id="agent-name">Error loading</span></p>
                <p><strong>Thumbprint:</strong> <span id="agent-thumbprint"></span></p>
                <p><strong>Last Communication:</strong> <span id="agent-lastcomm"></span></p>`;
    }
    return `<p><strong>Name:</strong> <span id="agent-name">${state.agent.node_name || ''}</span></p>
            <p><strong>Thumbprint:</strong> <span id="agent-thumbprint">${state.agent.thumbprint || state.agent.certificate_thumbprint || ''}</span></p>
            <p><strong>Last Communication:</strong> <span id="agent-lastcomm">${state.agent.last_communication || ''}</span></p>`;
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
        // Retire la colonne StartDate, ModuleVersion et ResourceId si présentes
        cols = cols.filter(c => c !== 'StartDate' && c !== 'ModuleVersion' && c !== 'ResourceId');
        let html = `<p><strong>${title}</strong></p><table style="width:100%" class="table table-bordered table-sm"><thead><tr>`;
        html += cols.map(c => `<th>${c}</th>`).join('');
        html += '</tr></thead><tbody>';
        for(const res of resources) {
            html += '<tr>' + cols.map(c => {
                if (c === 'ModuleName') {
                    let mod = res[c] || '';
                    let version = res['ModuleVersion'] || '';
                    if (mod) {
                        return `<td>${mod}${version ? '<span class="badge badge-info" style="font-size:90%; margin-left:6px;">' + version + '</span>' : ''}</td>`;
                    } else {
                        return `<td></td>`;
                    }
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
    $('#agent-info-block').html(renderAgentInfo());
    $('#reports-dropdown-block').html(renderReportsDropdown());
    $('#last-report-block').html(renderSelectedReport());
}

$(document).ready(function() {
    const agentId = getNodeIdFromUrl();
    // Ajoute des blocs pour le rendu
    $('.card-body').html(`
        <h3>Agent Information</h3>
        <div id="agent-info-block"></div>
        <hr>
        <h3>DSC Report</h3>
        <div id="reports-dropdown-block" style="margin-bottom:1em;"></div>
        <div id="last-report-block"></div>
    `);
    renderAll();
    if (!agentId) {
        state.error = 'Agent ID not found in URL';
        renderAll();
        return;
    }
    // Load agent info
    fetch(`/api/v1/agents/${agentId}`)
        .then(r => r.json())
        .then(agent => {
            state.agent = agent;
            renderAll();
        })
        .catch(() => {
            state.agent = null;
            renderAll();
        });
    // Charger la liste des rapports
    fetch(`/api/v1/agents/${agentId}/reports?operationtype=Initial`)
        .then(r => r.json())
        .then(reports => {
            state.reports = reports;
            // Par défaut, sélectionne le plus récent (premier)
            state.selectedReportIdx = 0;
            renderAll();
            // Déclenche le onchange pour charger le rapport sélectionné
            setTimeout(() => {
                $('#report-select').trigger('change');
            }, 0);
        })
        .catch(() => {
            state.reports = [];
            state.selectedReportIdx = null;
            state.selectedReport = null;
            renderAll();
        });

    // Gestion du changement de sélection de rapport (delegation)
    $(document).on('change', '#report-select', function() {
        const idx = parseInt($(this).val(), 10);
        state.selectedReportIdx = idx;
        const rep = state.reports[idx];
        if (rep && rep.job_id) {
            // Aller chercher le rapport détaillé
            fetch(`/api/v1/agents/${agentId}/reports/${rep.job_id}`)
                .then(r => r.json())
                .then(report => {
                    state.selectedReport = report;
                    renderAll();
                })
                .catch(() => {
                    state.selectedReport = null;
                    renderAll();
                });
        } else {
            state.selectedReport = rep || null;
            renderAll();
        }
    });
});
