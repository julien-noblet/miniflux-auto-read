local config = import '../config.libsonnet';

{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'miniflux-auto-read',
        rules: [
          {
            alert: 'MinifluxProcessingErrorsHigh',
            expr: 'sum(rate(miniflux_entries_processing_errors_total{%(minifluxAutoReadSelector)s}[5m])) > 0' % config._config,
            'for': '5m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'High processing error rate for Miniflux Auto Read',
              description: 'The error rate for Miniflux entry processing is currently {{ $value }}.',
            },
          },
          {
            alert: 'MinifluxHighLatency',
            expr: 'histogram_quantile(0.95, sum(rate(miniflux_api_duration_seconds_bucket{%(minifluxAutoReadSelector)s}[5m])) by (le)) > 2' % config._config,
            'for': '10m',
            labels: {
              severity: 'warning',
            },
            annotations: {
              summary: 'High Miniflux API latency',
              description: '95th percentile of Miniflux API calls is {{ $value }}s.',
            },
          },
        ],
      },
    ],
  },
}
