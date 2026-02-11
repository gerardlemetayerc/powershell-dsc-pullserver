$(function() {
    // Récupère l'ID depuis l'URL
    function getNameFromUrl() {
        var match = window.location.pathname.match(/configuration_model\/([^\/]+)/);
        return match ? decodeURIComponent(match[1]) : null;
    }
    var name = getNameFromUrl();
    if (!name) return;

    // Charge le détail de la configuration
    $.getJSON('/api/v1/configuration_models/' + encodeURIComponent(name) + '/detail', function(data) {
        var html = '';
        html += '<h3>' + data.name + '</h3>';
        html += '<p><b>Uploaded by:</b> ' + data.uploaded_by + ' | <b>Date:</b> ' + (data.upload_date ? new Date(data.upload_date).toLocaleString() : '') + '</p>';
        html += '<a class="btn btn-primary" href="/api/v1/configuration_models/' + encodeURIComponent(data.name) + '/download">Télécharger le MOF</a>';
        html += '<hr><h4>Historique des versions</h4>';
        html += '<ul>';
        data.versions.forEach(function(v) {
            html += '<li><a href="/web/configuration_model/' + encodeURIComponent(v.name) + '">' + v.name + '</a> (' + (v.upload_date ? new Date(v.upload_date).toLocaleString() : '') + ')</li>';
        });
        html += '</ul>';
        // Affichage des noeuds liés
        html += '<hr><h4>Noeuds liés à cette configuration</h4>';
        if(data.linked_nodes && data.linked_nodes.length) {
            html += '<div class="table-responsive"><table class="table table-sm table-bordered"><thead><tr><th>Agent ID</th><th>Nom du noeud</th><th>État</th><th>Dernière communication</th></tr></thead><tbody>';
            data.linked_nodes.forEach(function(n) {
                html += '<tr>';
                html += '<td>' + n.agent_id + '</td>';
                html += '<td>' + (n.node_name || '') + '</td>';
                html += '<td>' + (n.state || '') + '</td>';
                html += '<td>' + (n.last_communication ? new Date(n.last_communication).toLocaleString() : '') + '</td>';
                html += '</tr>';
            });
            html += '</tbody></table></div>';
        } else {
            html += '<p>Aucun noeud lié.</p>';
        }

        // Affichage des modules utilisés
        html += '<hr><h4>Modules utilisés</h4>';
        if(data.modules && data.modules.length) {
            html += '<div class="table-responsive"><table class="table table-sm table-bordered"><thead><tr><th>Nom</th><th>Version requise</th><th>Version disponible</th><th></th></tr></thead><tbody>';
            data.modules.forEach(function(m, idx) {
                var modId = 'modver_' + idx;
                html += '<tr>';
                html += '<td>' + (m.name || '') + '</td>';
                html += '<td>' + (m.version || '') + '</td>';
                html += '<td id="' + modId + '"><span class="text-muted">Vérification...</span></td>';
                html += '<td>' + (m.dependencies && m.dependencies.length ? m.dependencies.join(', ') : '') + '</td>';
                html += '</tr>';
            });
            html += '</tbody></table></div>';
        } else {
            html += '<p>Aucun module détecté.</p>';
        }
        $('#config-detail-container').html(html);

        // Vérification des versions de modules
        if(data.modules && data.modules.length) {
            data.modules.forEach(function(m, idx) {
                var modId = 'modver_' + idx;
                if(!m.name) return;
                var url = '/api/v1/modules/' + encodeURIComponent(m.name);
                if(m.version) url += '?latest=1';
                $.getJSON(url, function(resp) {
                    // resp doit contenir {available: bool, version: string}
                    var cell = $('#' + modId);
                    if(resp && resp.available) {
                        var ok = true;
                        if(m.version && resp.version) {
                            // Compare les versions (simple split)
                            var req = m.version.split('.').map(Number);
                            var got = resp.version.split('.').map(Number);
                            for(var i=0; i<req.length; ++i) {
                                if((got[i]||0) > req[i]) break;
                                if((got[i]||0) < req[i]) { ok = false; break; }
                            }
                        }
                        cell.html('<span style="color:'+(ok?'green':'red')+'">' + resp.version + '</span>');
                    } else {
                        $('#' + modId).html('<span style="color:red">Non disponible</span>');
                    }
                }).fail(function() {
                    $('#' + modId).html('<span style="color:red">Erreur</span>');
                });
            });
        }
    });
});
