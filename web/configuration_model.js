$(function() {
    // Affiche la modal d'upload et charge les propriétés
    $('#btn-upload-config-model').on('click', function() {
        $.getJSON('/api/v1/properties', function(props) {
            var options = '';
            props.forEach(function(p) {
                options += '<option value="' + p.name + '">' + p.name + ' (prio ' + p.priority + ')</option>';
            });
            $('#property-select').html(options);
            $('#modal-upload-config-model').modal('show');
        });
    });

    // Soumission du formulaire d'upload
    $('#form-upload-config-model').on('submit', function(e) {
        e.preventDefault();
        var formData = new FormData(this);
        $.ajax({
            url: '/api/v1/configuration_models',
            type: 'POST',
            data: formData,
            processData: false,
            contentType: false,
            success: function() {
                $('#modal-upload-config-model').modal('hide');
                loadConfigModels();
            },
            error: function(xhr) {
                alert('Erreur upload: ' + xhr.responseText);
            }
        });
    });

    // Chargement du tableau
    function loadConfigModels() {
        $.getJSON('/api/v1/configuration_models', function(data) {
            var tbody = '';
            data.forEach(function(model) {
                tbody += '<tr>' +
                    '<td>' + model.id + '</td>' +
                    '<td>' + model.property + '</td>' +
                    '<td>' + model.value + '</td>' +
                    '<td>' + model.upload_date + '</td>' +
                    '<td>' +
                        '<button class="btn btn-danger btn-sm delete-config-model" data-id="' + model.id + '">Supprimer</button>' +
                    '</td>' +
                '</tr>';
            });
            $('#config-models-table tbody').html(tbody);
        });
    }

    // Suppression
    $('#config-models-table').on('click', '.delete-config-model', function() {
        var id = $(this).data('id');
        if (confirm('Supprimer ce modèle ?')) {
            $.ajax({
                url: '/api/v1/configuration_models?id=' + id,
                type: 'DELETE',
                success: function() {
                    loadConfigModels();
                },
                error: function(xhr) {
                    alert('Erreur suppression: ' + xhr.responseText);
                }
            });
        }
    });

    // Initial load
    loadConfigModels();
});
