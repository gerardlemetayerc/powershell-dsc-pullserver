
$(document).ready(function() {
        // Affiche le nombre total de configurations
        $.getJSON('/api/v1/configuration_models?count=1', function(data) {
            if (data && typeof data.count !== 'undefined') {
                $('#total-configurations').text(data.count);
            }
        });
    var dataTable = $('#agents-table').DataTable({
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
                    if (row.state) {
                        switch (row.state.toLowerCase()) {
                            case 'ok':
                                return `<span class="badge bg-success" style="min-width:60px;">OK</span>`;
                            case 'waiting_for_registration':
                                return `<span class="badge bg-secondary" style="min-width:100px;">Pending Enroll</span>`;
                            case 'failure':
                                return `<span class="badge bg-danger" style="min-width:60px;">Failed</span>`;
                            default:
                                return `<span class="badge bg-info" style="min-width:80px;">${row.state}</span>`;
                        }
                    } else {
                        return `<span class="badge bg-info" style="min-width:80px;">Unknown</span>`;
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
        ]
    });
    $(window).on('resize', function() {
        $('#agents-table').DataTable().columns.adjust().responsive.recalc();
    });
    $.getJSON('/api/v1/agents', function(agents) {
        // Affiche le nombre total d'agents
        $('#total-agents').text(agents.length);
        // Calcule le nombre d'agents par état principal
        let ok = 0, err = 0, pendingEnroll = 0, pendingApply = 0;
        agents.forEach(a => {
            if (!a.state) return;
            switch (a.state.toLowerCase()) {
                case 'ok':
                    ok++;
                    break;
                case 'failure':
                    err++;
                    break;
                case 'waiting_for_registration':
                    pendingEnroll++;
                    break;
                case 'pending_apply':
                    pendingApply++;
                    break;
                default:
                    // ignore or handle as needed
                    break;
            }
        });
        // Affiche le camembert avec 4 catégories
        const ctx = document.getElementById('agents-pie').getContext('2d');
        const chart = new Chart(ctx, {
            type: 'pie',
            data: {
                labels: ['OK', 'Failed', 'Pending Enroll', 'Pending Apply'],
                datasets: [{
                    data: [ok, err, pendingEnroll, pendingApply],
                    backgroundColor: ['#28a745', '#dc3545', '#6c757d', '#ffc107'],
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
            let showPendingEnroll = $('#legend-pending-enroll').hasClass('active');
            let showPendingApply = $('#legend-pending-apply').hasClass('active');
            let newData = [showOk ? ok : 0, showErr ? err : 0, showPendingEnroll ? pendingEnroll : 0, showPendingApply ? pendingApply : 0];
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
        $('#legend-pending-enroll').on('click', function() {
            $(this).toggleClass('active');
            updatePieVisibility();
        });
        $('#legend-pending-apply').on('click', function() {
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
    $('#preenroll-btn').click(function() {
        $('#preenrollModal').modal('show');
    });
    $('#preenroll-submit').click(function() {
        var nodeName = $('#preenroll-node-name').val();
        if (!nodeName) {
            $('#preenroll-error').text('Node name is required');
            return;
        }
        $('#preenroll-error').text('');
        $.ajax({
            url: '/api/v1/agents/preenroll',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ node_name: nodeName }),
            // header Authorization ajouté globalement par jwt_ajax.js
            success: function(data) {
                $('#preenrollModal').modal('hide');
                $('#preenroll-node-name').val('');
                // Optionally refresh the agents table
                dataTable.ajax.reload(null, false);
            },
            error: function(xhr) {
                $('#preenroll-error').text(xhr.responseText || 'Error');
            }
        });
    });
});