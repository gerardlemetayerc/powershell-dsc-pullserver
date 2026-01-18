$(function() {
    // Affiche la modal d'upload et charge les propriétés
    $('#btn-upload-config-model').on('click', function() {
        $('#modal-upload-config-model').modal('show');
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
    function formatDate(dateStr) {
        if (!dateStr || dateStr === "0001-01-01T00:00:00Z") return "";
        var d = new Date(dateStr);
        if (isNaN(d.getTime())) return dateStr;
        return d.toLocaleString();
    }

    function loadConfigModels() {
        $.getJSON('/api/v1/configuration_models', function(data) {
            if ($.fn.DataTable.isDataTable('#config-models-table')) {
                $('#config-models-table').DataTable().clear().destroy();
            }
            var tbody = '';
            data.forEach(function(model) {
                var lastUsage = formatDate(model.last_usage);
                var uploadDate = formatDate(model.upload_date);
                tbody += '<tr>' +
                    '<td>' + model.id + '</td>' +
                    '<td>' + model.name + '</td>' +
                    '<td>' + model.uploaded_by + '</td>' +
                    '<td>' + uploadDate + '</td>' +
                    '<td>' + lastUsage + '</td>' +
                    '<td>' +
                        '<button class="btn btn-danger btn-sm delete-config-model" data-id="' + model.id + '">Supprimer</button>' +
                    '</td>' +
                '</tr>';
            });
            $('#config-models-table tbody').html(tbody);
            $('#config-models-table').DataTable({
                language: {
                    url: '//cdn.datatables.net/plug-ins/1.13.4/i18n/fr-FR.json'
                },
                order: [[0, 'desc']]
            });
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
