$(document).ready(function() {
    if(window.location.pathname !== '/web/admin/audit') {
        return;
    }

    function selectedValues(selector) {
        const values = $(selector).val();
        return Array.isArray(values) ? values : [];
    }

    function loadFilterOptions() {
        $.get('/api/v1/audit/filters')
            .done(function(resp) {
                const users = (resp && Array.isArray(resp.users)) ? resp.users : [];
                const actions = (resp && Array.isArray(resp.actions)) ? resp.actions : [];

                const $userSelect = $('#audit-filter-users');
                const $actionSelect = $('#audit-filter-actions');
                $userSelect.empty();
                $actionSelect.empty();

                users.forEach(function(user) {
                    $userSelect.append($('<option>', { value: user, text: user }));
                });
                actions.forEach(function(action) {
                    $actionSelect.append($('<option>', { value: action, text: action }));
                });
            });
    }

    const table = $('#audit-table').DataTable({
        processing: true,
        serverSide: true,
        ordering: true,
        autoWidth: false,
        pageLength: 20,
        lengthMenu: [20, 50, 100],
        order: [[4, 'desc']],
        ajax: function(data, callback) {
            const orderColumn = (data.order && data.order.length > 0) ? data.order[0].column : 4;
            const orderDir = (data.order && data.order.length > 0) ? data.order[0].dir : 'desc';
            $.get('/api/v1/audit', {
                draw: data.draw,
                start: data.start,
                length: data.length,
                'order[0][column]': orderColumn,
                'order[0][dir]': orderDir,
                orderColumn: orderColumn,
                orderDir: orderDir,
                users: selectedValues('#audit-filter-users').join(','),
                actions: selectedValues('#audit-filter-actions').join(','),
                date_from: $('#audit-filter-date-from').val() || '',
                date_to: $('#audit-filter-date-to').val() || '',
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

    $('#audit-filter-users, #audit-filter-actions, #audit-filter-date-from, #audit-filter-date-to').on('change', function() {
        table.ajax.reload();
    });

    $('#audit-filter-reset').on('click', function() {
        $('#audit-filter-users').val([]);
        $('#audit-filter-actions').val([]);
        $('#audit-filter-date-from').val('');
        $('#audit-filter-date-to').val('');
        table.search('');
        table.ajax.reload();
    });

    loadFilterOptions();

    $('#audit-export-csv').on('click', function() {
        const search = table.search() || '';
        const users = selectedValues('#audit-filter-users').join(',');
        const actions = selectedValues('#audit-filter-actions').join(',');
        const dateFrom = $('#audit-filter-date-from').val() || '';
        const dateTo = $('#audit-filter-date-to').val() || '';
        const url = '/api/v1/audit/export?search=' + encodeURIComponent(search)
            + '&users=' + encodeURIComponent(users)
            + '&actions=' + encodeURIComponent(actions)
            + '&date_from=' + encodeURIComponent(dateFrom)
            + '&date_to=' + encodeURIComponent(dateTo);
        window.location.href = url;
    });
});
