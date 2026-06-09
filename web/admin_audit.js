$(document).ready(function() {
    if(window.location.pathname !== '/web/admin/audit') {
        return;
    }

    const table = $('#audit-table').DataTable({
        processing: true,
        serverSide: true,
        autoWidth: false,
        pageLength: 20,
        lengthMenu: [20, 50, 100],
        order: [[4, 'desc']],
        ajax: function(data, callback) {
            $.get('/api/v1/audit', {
                draw: data.draw,
                start: data.start,
                length: data.length,
                'search[value]': (data.search && data.search.value) ? data.search.value : ''
            })
                .done(function(resp) {
                    callback({
                        draw: resp.draw || data.draw,
                        recordsTotal: resp.recordsTotal || resp.total || 0,
                        recordsFiltered: resp.recordsFiltered || resp.filtered || resp.total || 0,
                        data: resp.data || resp.items || []
                    });
                })
                .fail(function() {
                    callback({
                        draw: data.draw,
                        recordsTotal: 0,
                        recordsFiltered: 0,
                        data: []
                    });
                });
        },
        columns: [
            { data: 'UserEmail', defaultContent: '' },
            { data: 'Action', defaultContent: '' },
            { data: 'Target', defaultContent: '' },
            { data: 'Details', defaultContent: '' },
            { data: 'CreatedAt', defaultContent: '' }
        ]
    });

    $('#audit-export-csv').on('click', function() {
        const search = table.search() || '';
        const url = '/api/v1/audit/export?search=' + encodeURIComponent(search);
        window.location.href = url;
    });
});
