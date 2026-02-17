# Development

## kcm

```bash
  git clone https://github.com/k0rdent/kcm.git
  cd kcm
  make cli-install
  make dev-apply
```

## vlogxy

1. Fork the [k0rdent/vlogxy](https://github.com/k0rdent/vlogxy) repository to your own account, e.g. `https://github.com/YOUR_USERNAME/vlogxy`.

2. Run the following commands:

```bash
  cd ..
  git clone git@github.com:YOUR_USERNAME/vlogxy.git
  cd kof

  make cli-install
```

3. Deploy the `vlogxy` chart to your local management cluster:

```bash
make dev-deploy
```
