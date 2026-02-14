$(function() {
    const $form = $('#saml-config-form');
    const $status = $('#saml-config-status');

    // Affiche le nom du fichier sélectionné pour chaque input file
    $('#sp_key_upload').on('change', function() {
        var fileName = $(this).val().split('\\').pop();
        $(this).next('.custom-file-label').addClass('selected').text(fileName || 'Upload Private Key (.key/.pem)');
    });
    $('#sp_cert_upload').on('change', function() {
        var fileName = $(this).val().split('\\').pop();
        $(this).next('.custom-file-label').addClass('selected').text(fileName || 'Upload Certificate (.crt/.pem)');
    });

    function toggleFields(enabled) {
        const $fields = $form.find('.saml-fields');
        if (enabled) {
            $fields.show();
        } else {
            $fields.hide();
        }
    }

    let claims = {};
    const baseClaims = ["email", "givenname", "sn"];
    function renderClaimsTable() {
        // Remplit les inputs fixes
        baseClaims.forEach(field => {
            $(`#claim-${field}-uri`).val(claims[field] || "");
        });
        // Claims additionnels dynamiques
        const $tbody = $('#claims-table tbody');
        $tbody.find('tr').slice(3).remove(); // Garde les 3 premiers (fixes)
        Object.entries(claims).forEach(([field, uri]) => {
            if (!baseClaims.includes(field)) {
                $tbody.append(`<tr><td>${field}</td><td>${uri}</td><td><button type="button" class="btn btn-danger btn-sm remove-claim" data-field="${field}"><i class="fas fa-trash"></i></button></td></tr>`);
            }
        });
    }

    function loadConfig() {
        $.get('/api/v1/saml/config')
        .done(function(cfg) {
            $form.find('[name="enabled"]').prop('checked', !!cfg.enabled);
            $form.find('[name="entity_id"]').val(cfg.entity_id || '');
            $form.find('[name="sp_cert_file"]').val(cfg.sp_cert_file || '');
            $form.find('[name="sp_key_file"]').val(cfg.sp_key_file || '');
            $form.find('[name="idp_metadata_url"]').val(cfg.idp_metadata_url || '');
            claims = (cfg.user_mapping && typeof cfg.user_mapping === 'object') ? {...cfg.user_mapping} : {};
            // Group mapping
            const group = (cfg.group_mapping && typeof cfg.group_mapping === 'object') ? cfg.group_mapping : {};
            $('#group-attribute').val(group.attribute || '');
            $('#group-admin-value').val(group.admin_value || '');
            $('#group-user-value').val(group.user_value || '');
            renderClaimsTable();
            toggleFields(!!cfg.enabled);
        })
        .fail(function(xhr) {
            $status.text('Failed to load config.').css('color', 'red');
        });
    }

    $form.find('[name="enabled"]').on('change', function() {
        toggleFields($(this).prop('checked'));
    });

    // Masquer les champs au chargement
    toggleFields(false);



    $(document).on('input', '.claim-uri-input', function() {
        const field = $(this).data('claim');
        claims[field] = $(this).val();
    });
    $form.on('submit', function(e) {
        e.preventDefault();
        const group_mapping = {
            attribute: $('#group-attribute').val(),
            admin_value: $('#group-admin-value').val(),
            user_value: $('#group-user-value').val()
        };
        const data = {
            enabled: $form.find('[name="enabled"]').prop('checked'),
            entity_id: $form.find('[name="entity_id"]').val(),
            sp_cert_file: $form.find('[name="sp_cert_file"]').val(),
            sp_key_file: $form.find('[name="sp_key_file"]').val(),
            idp_metadata_url: $form.find('[name="idp_metadata_url"]').val(),
            user_mapping: claims,
            group_mapping: group_mapping
        };
        $.ajax({
            url: '/api/v1/saml/config',
            method: 'PUT',
            contentType: 'application/json',
            data: JSON.stringify(data)
        })
        .done(function() {
            $status.text('Configuration updated!').css('color', 'green');
        })
        .fail(function(xhr) {
            $status.text('Update failed.').css('color', 'red');
        });
    });

    // Ajout/suppression de claims additionnels désactivé

    // Upload SP key/cert
    $('#upload-sp-keycert').on('click', function() {
        const keyFile = $('#sp_key_upload')[0].files[0];
        const certFile = $('#sp_cert_upload')[0].files[0];
        const $uploadStatus = $('#sp-upload-status');
        if (!keyFile || !certFile) {
            $uploadStatus.text('Please select both key and certificate files.').css('color', 'red');
            return;
        }
        var formData = new FormData();
        formData.append('sp_key', keyFile);
        formData.append('sp_cert', certFile);
        $.ajax({
            url: '/api/v1/saml/upload_sp_keycert',
            method: 'POST',
            data: formData,
            processData: false,
            contentType: false,
            success: function() {
                $uploadStatus.text('Key/certificate uploaded and validated!').css('color', 'green');
            },
            error: function(xhr) {
                let msg = 'Upload failed.';
                if(xhr.responseJSON && xhr.responseJSON.error) msg = xhr.responseJSON.error;
                $uploadStatus.text(msg).css('color', 'red');
            }
        });
    });

    loadConfig();
});
