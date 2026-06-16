$(function() {
    const $form = $('#provisioning-config-form');
    const $status = $('#provisioning-config-status');
    const $secretPresence = $('#secret-presence-text');

    function showStatus(text, color) {
        $status.text(text).css('color', color);
    }

    function fillForm(cfg) {
        $form.find('#enabled').prop('checked', !!cfg.enabled);
        $form.find('#provider').val(cfg.provider || 'github');
        $form.find('#api_base_url').val(cfg.api_base_url || '');
        $form.find('#project_path').val(cfg.project_path || '');
        $form.find('#workflow_id').val(cfg.workflow_id || '');
        $form.find('#pipeline_ref').val(cfg.pipeline_ref || 'main');
        $form.find('#timeout_seconds').val(cfg.timeout_seconds || 30);
        $form.find('#secret_token').val('');
        $secretPresence.text(cfg.has_secret ? 'A secret is already configured.' : 'No secret configured yet.');
    }

    function loadConfig() {
        $.get('/api/v1/provisioning/pipeline/config')
            .done(function(cfg) {
                fillForm(cfg || {});
            })
            .fail(function() {
                showStatus('Failed to load provisioning config.', 'red');
            });
    }

    $form.on('submit', function(e) {
        e.preventDefault();

        const payload = {
            enabled: $form.find('#enabled').prop('checked'),
            provider: ($form.find('#provider').val() || '').trim(),
            api_base_url: ($form.find('#api_base_url').val() || '').trim(),
            project_path: ($form.find('#project_path').val() || '').trim(),
            workflow_id: ($form.find('#workflow_id').val() || '').trim(),
            pipeline_ref: ($form.find('#pipeline_ref').val() || '').trim(),
            timeout_seconds: parseInt($form.find('#timeout_seconds').val(), 10) || 30
        };

        const secret = ($form.find('#secret_token').val() || '').trim();
        if (secret !== '') {
            payload.secret_token = secret;
        }

        $.ajax({
            url: '/api/v1/provisioning/pipeline/config',
            method: 'PUT',
            contentType: 'application/json',
            data: JSON.stringify(payload)
        })
            .done(function(resp) {
                fillForm(resp || payload);
                showStatus('Provisioning configuration updated.', 'green');
            })
            .fail(function(xhr) {
                const msg = (xhr.responseJSON && xhr.responseJSON.error) ? xhr.responseJSON.error : 'Failed to update provisioning config.';
                showStatus(msg, 'red');
            });
    });

    loadConfig();
});
