$(function() {
    // Get the ID from the URL
    function getNameFromUrl() {
        var match = window.location.pathname.match(/configuration_model\/([^\/]+)/);
        return match ? decodeURIComponent(match[1]) : null;
    }

    function formatDate(value) {
        if (!value) return '';
        var dt = new Date(value);
        return isNaN(dt.getTime()) ? value : dt.toLocaleString();
    }

    function renderStatusBadge(status) {
        if (!status) return '<span class="badge badge-secondary">No run yet</span>';
        var normalized = String(status).toLowerCase();
        var badgeClass = 'badge-secondary';
        if (normalized === 'success' || normalized === 'ok') {
            badgeClass = 'badge-success';
        } else if (normalized === 'failure' || normalized === 'failed' || normalized === 'error') {
            badgeClass = 'badge-danger';
        } else if (normalized === 'pending_apply' || normalized === 'getconfiguration') {
            badgeClass = 'badge-warning';
        }
        return '<span class="badge ' + badgeClass + '">' + status + '</span>';
    }

    function destroyDataTable(selector) {
        if ($.fn.DataTable.isDataTable(selector)) {
            $(selector).DataTable().clear().destroy();
        }
    }

    function initDetailDataTables() {
        ['#config-version-history-table', '#config-linked-nodes-table', '#config-used-modules-table'].forEach(function(selector) {
            if ($(selector).length) {
                destroyDataTable(selector);
                $(selector).DataTable({
                    pageLength: 10,
                    order: []
                });
            }
        });
    }

    var name = getNameFromUrl();
    if (!name) return;

    // Load configuration details
    $.getJSON('/api/v1/configuration_models/' + encodeURIComponent(name) + '/detail', function(data) {
        var html = '';
        html += '<h3>' + data.name + '</h3>';
        html += '<p><b>Uploaded by:</b> ' + data.uploaded_by + ' | <b>Date:</b> ' + formatDate(data.upload_date) + '</p>';
        html += '<button class="btn btn-primary" id="download-mof-btn">Download MOF</button>';
        html += '<hr><h4>Version history</h4>';
        if(data.versions && data.versions.length) {
            html += '<div class="table-responsive"><table id="config-version-history-table" class="table table-sm table-bordered"><thead><tr><th>Name</th><th>Current</th><th>Uploaded at</th><th>Uploaded by</th></tr></thead><tbody>';
            data.versions.forEach(function(v) {
                var isCurrent = (v.name === data.name);
                html += '<tr>';
                html += '<td><a href="/web/configuration_model/' + encodeURIComponent(v.name) + '">' + v.name + '</a></td>';
                html += '<td>' + (isCurrent ? '<span class="badge badge-primary">Current</span>' : '') + '</td>';
                html += '<td>' + formatDate(v.upload_date) + '</td>';
                html += '<td>' + (v.uploaded_by || '') + '</td>';
                html += '</tr>';
            });
            html += '</tbody></table></div>';
        } else {
            html += '<p>No version history.</p>';
        }
        // Linked nodes display
        html += '<hr><h4>Nodes linked to this configuration</h4>';
        if(data.linked_nodes && data.linked_nodes.length) {
            html += '<div class="table-responsive"><table id="config-linked-nodes-table" class="table table-sm table-bordered"><thead><tr><th>Agent ID</th><th>Node name</th><th>Last execution status</th><th>Last execution at</th><th>Last communication</th></tr></thead><tbody>';
            data.linked_nodes.forEach(function(n) {
                html += '<tr>';
                html += '<td>' + n.agent_id + '</td>';
                html += '<td>' + (n.node_name || '') + '</td>';
                html += '<td>' + renderStatusBadge(n.last_execution_status) + '</td>';
                html += '<td>' + formatDate(n.last_execution_at) + '</td>';
                html += '<td>' + formatDate(n.last_communication) + '</td>';
                html += '</tr>';
            });
            html += '</tbody></table></div>';
        } else {
            html += '<p>No linked nodes.</p>';
        }

        // Used modules display
        html += '<hr><h4>Used modules</h4>';
        if(data.modules && data.modules.length) {
            html += '<div class="table-responsive"><table id="config-used-modules-table" class="table table-sm table-bordered"><thead><tr><th>Name</th><th>Required version</th><th>Available version</th></tr></thead><tbody>';
            data.modules.forEach(function(m, idx) {
                var modId = 'modver_' + idx;
                html += '<tr>';
                html += '<td>' + (m.name || '') + '</td>';
                html += '<td>' + (m.version || '') + '</td>';
                html += '<td id="' + modId + '"><span class="text-muted">Checking...</span></td>';
                html += '</tr>';
            });
            html += '</tbody></table></div>';
        } else {
            html += '<p>No modules detected.</p>';
        }
        $('#config-detail-container').html(html);
        initDetailDataTables();
        // MOF download with auth
        $('#download-mof-btn').on('click', function() {
            var url = '/api/v1/configuration_models/' + encodeURIComponent(data.id) + '/download';
            var xhr = new XMLHttpRequest();
            xhr.open('GET', url, true);
            xhr.responseType = 'blob';
            xhr.onload = function() {
                if(xhr.status === 200) {
                    var blob = xhr.response;
                    var link = document.createElement('a');
                    link.href = window.URL.createObjectURL(blob);
                    link.download = data.name + '.mof';
                    document.body.appendChild(link);
                    link.click();
                    document.body.removeChild(link);
                } else {
                    alert('Download failed: ' + xhr.status);
                }
            };
            xhr.onerror = function() { alert('Download error'); };
            xhr.send();
        });

        // Module version check
        if(data.modules && data.modules.length) {
            data.modules.forEach(function(m, idx) {
                var modId = 'modver_' + idx;
                if(!m.name) return;
                var url = '/api/v1/modules/' + encodeURIComponent(m.name);
                if(m.version) url += '?latest=1';
                $.getJSON(url, function(resp) {
                    // resp should contain {available: bool, version: string}
                    var cell = $('#' + modId);
                    if(resp && resp.available) {
                        var ok = true;
                        if(m.version && resp.version) {
                            // Simple version split/compare
                            var req = m.version.split('.').map(Number);
                            var got = resp.version.split('.').map(Number);
                            for(var i=0; i<req.length; ++i) {
                                if((got[i]||0) > req[i]) break;
                                if((got[i]||0) < req[i]) { ok = false; break; }
                            }
                        }
                        cell.html('<span style="color:'+(ok?'green':'red')+'">' + resp.version + '</span>');
                    } else {
                        $('#' + modId).html('<span style="color:red">Not available</span>');
                    }
                }).fail(function() {
                    $('#' + modId).html('<span style="color:red">Error</span>');
                });
            });
        }
    });
});
