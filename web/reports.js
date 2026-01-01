
$(document).ready(function() {
    $('#agents-table').DataTable({
        ajax: {
            url: '/api/v1/agents',
            dataSrc: ''
        },
        columns: [
            { data: 'node_name', title: 'Nom' },
            { data: 'last_communication', title: 'Dernière communication' },
            { data: 'certificate_thumbprint', title: 'Thumbprint' },
            {
                data: 'agent_id',
                title: 'Détail',
                render: function(data, type, row) {
                    return `<a class='btn btn-info btn-sm' href="/web/node/${data}">Voir</a>`;
                }
            }
        ],
        language: {
            url: '/web/fr-FR.json'
        }
    });
});

// La fonction viewAgentDetail n'est plus utilisée, la redirection se fait via le lien.
