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
                $('#enable_release_check').prop('checked', !!cfg.enable_release_check);
                $('#release_check_interval_mins').val(cfg.release_check_interval_mins || 1440);
            })
            .fail(function() {
                showStatus('Unable to load configuration.', 'red');
            });
    }

    $form.on('submit', function(e) {
        e.preventDefault();
        const payload = {
            enable_report_auto_cleanup: $('#enable_report_auto_cleanup').prop('checked'),
            report_retention_days: parseInt($('#report_retention_days').val(), 10),
            report_cleanup_interval_mins: parseInt($('#report_cleanup_interval_mins').val(), 10),
            enable_release_check: $('#enable_release_check').prop('checked'),
            release_check_interval_mins: parseInt($('#release_check_interval_mins').val(), 10)
        };

        $.ajax({
            url: '/api/v1/scheduler/config',
            method: 'PUT',
            contentType: 'application/json',
            data: JSON.stringify(payload)
        })
            .done(function() {
                showStatus('Configuration saved.', 'green');
            })
            .fail(function(xhr) {
                const msg = (xhr.responseJSON && xhr.responseJSON.error) ? xhr.responseJSON.error : 'Failed to save configuration.';
                showStatus(msg, 'red');
            });
    });

    $('#run-cleanup-now').on('click', function() {
        $.ajax({
            url: '/api/v1/scheduler/cleanup/run',
            method: 'POST'
        })
            .done(function(resp) {
                showStatus('Cleanup completed. Deleted reports: ' + (resp.deleted || 0), 'green');
            })
            .fail(function(xhr) {
                const msg = xhr.responseText || 'Manual cleanup failed.';
                showStatus(msg, 'red');
            });
    });

    $('#run-release-check-now').on('click', function() {
        $.ajax({
            url: '/api/v1/scheduler/release-check/run',
            method: 'POST'
        })
            .done(function(resp) {
                if (resp.update_available) {
                    showStatus('Update available: ' + resp.latest_release + ' (current: ' + resp.current_version + ')', 'orange');
                } else {
                    showStatus('Up to date: ' + (resp.current_version || '-'), 'green');
                }
            })
            .fail(function(xhr) {
                const msg = xhr.responseText || 'Release check failed.';
                showStatus(msg, 'red');
            });
    });

    loadConfig();
});
