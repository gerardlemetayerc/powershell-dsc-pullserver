$(function() {
    const $form = $('#scheduler-config-form');
    const $status = $('#scheduler-config-status');

    function showStatus(text, color) {
        $status.text(text).css('color', color);
    }

    function loadConfig() {
        $.get('/api/v1/scheduler/config')
            .done(function(cfg) {
                $('#enable_report_auto_cleanup').prop('checked', !!cfg.enable_report_auto_cleanup);
                $('#report_retention_days').val(cfg.report_retention_days || 90);
                $('#report_cleanup_interval_mins').val(cfg.report_cleanup_interval_mins || 1440);
            })
            .fail(function() {
                showStatus('Impossible de charger la configuration.', 'red');
            });
    }

    $form.on('submit', function(e) {
        e.preventDefault();
        const payload = {
            enable_report_auto_cleanup: $('#enable_report_auto_cleanup').prop('checked'),
            report_retention_days: parseInt($('#report_retention_days').val(), 10),
            report_cleanup_interval_mins: parseInt($('#report_cleanup_interval_mins').val(), 10)
        };

        $.ajax({
            url: '/api/v1/scheduler/config',
            method: 'PUT',
            contentType: 'application/json',
            data: JSON.stringify(payload)
        })
            .done(function() {
                showStatus('Configuration sauvegardee.', 'green');
            })
            .fail(function(xhr) {
                const msg = (xhr.responseJSON && xhr.responseJSON.error) ? xhr.responseJSON.error : 'Echec de la sauvegarde.';
                showStatus(msg, 'red');
            });
    });

    $('#run-cleanup-now').on('click', function() {
        $.ajax({
            url: '/api/v1/scheduler/cleanup/run',
            method: 'POST'
        })
            .done(function(resp) {
                showStatus('Nettoyage termine. Rapports supprimes: ' + (resp.deleted || 0), 'green');
            })
            .fail(function(xhr) {
                const msg = xhr.responseText || 'Echec du nettoyage manuel.';
                showStatus(msg, 'red');
            });
    });

    loadConfig();
});
