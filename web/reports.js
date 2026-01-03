
$(document).ready(function() {
    $('#agents-table').DataTable({
        ajax: {
            url: '/api/v1/agents',
            dataSrc: ''
        },
        columns: [
            { data: 'node_name', title: 'Name' },
            { data: 'last_communication', title: 'Last Communication' },
            { data: 'certificate_thumbprint', title: 'Thumbprint' },
            {
                data: 'has_error_last_report',
                title: 'Status',
                className: 'text-center',
                render: function(data, type, row) {
                    if (data === false || data === 0 || data === 'false') {
                        return `<span class="badge bg-success" style="min-width:60px;">OK</span>`;
                    } else {
                        return `<span class="badge bg-danger" style="min-width:60px;">Error</span>`;
                    }
                }
            },
            {
                data: 'agent_id',
                title: 'Detail',
                render: function(data, type, row) {
                    return `<a class='btn btn-info btn-sm' href="/web/node/${data}">View</a>`;
                }
            }
        ],
        language: {
            url: '/web/fr-FR.json'
        }
    });
    $(window).on('resize', function() {
        $('#agents-table').DataTable().columns.adjust().responsive.recalc();
    });
    $.getJSON('/api/v1/agents', function(agents) {
        // Affiche le nombre total d'agents
        $('#total-agents').text(agents.length);
        // Calcule le nombre d'agents OK/Erreur
        let ok = 0, err = 0;
        agents.forEach(a => {
            if (a.has_error_last_report) err++; else ok++;
        });
        // Affiche le camembert
        const ctx = document.getElementById('agents-pie').getContext('2d');
        const chart = new Chart(ctx, {
            type: 'pie',
            data: {
                labels: ['OK', 'Erreur'],
                datasets: [{
                    data: [ok, err],
                    backgroundColor: ['#28a745', '#dc3545'],
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: { display: false }
                }
            }
        });

        // Gestion sélection légende custom
        function updatePieVisibility() {
            let showOk = $('#legend-ok').hasClass('active');
            let showErr = $('#legend-err').hasClass('active');
            let newData = [showOk ? ok : 0, showErr ? err : 0];
            chart.data.datasets[0].data = newData;
            chart.update();
        }
        $('#legend-ok').on('click', function() {
            $(this).toggleClass('active');
            updatePieVisibility();
        });
        $('#legend-err').on('click', function() {
            $(this).toggleClass('active');
            updatePieVisibility();
        });
        // Style visuel pour .active
        $('<style>.legend-selectable{opacity:0.6;} .legend-selectable.active{font-weight:bold;opacity:1;text-decoration:underline;}</style>').appendTo('head');
    });

    // Affiche le nombre total de modules
    $.getJSON('/api/v1/modules?count=1', function(data) {
        if (data && typeof data.count !== 'undefined') {
            $('#total-modules').text(data.count);
        }
    });
});

// La fonction viewAgentDetail n'est plus utilisée, la redirection se fait via le lien.
