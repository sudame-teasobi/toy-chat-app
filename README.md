# toy-chat-app

## セットアップ

### kubernetes のセットアップ

```bash
kubectl create ns toy-chat-app
```

#### strimzi (Kafka operator) のインストール

```bash
kubectl -n toy-chat-app apply -f 'https://strimzi.io/install/latest?namespace=toy-chat-app'

# 不要になった場合は:
kubectl -n toy-chat-app delete -f 'https://strimzi.io/install/latest?namespace=toy-chat-app'
```