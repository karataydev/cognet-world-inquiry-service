# 🌍 Cognate World Inquiry Service

A simple service to explore word cognates across different languages. Built with Go and Redis.

## 📚 What it does

- Imports cognate data from TSV files
- Provides word suggestions as you type
- Shows related words across languages

## ✨ Acknowledgments

This project uses the CogNet dataset. Special thanks to:
- Khuyagbaatar Batsuren
- Gábor Bella
- Fausto Giunchiglia

For their work on CogNet. If you use this service or the dataset, please cite their paper:

```bibtex
@inproceedings{batsuren2019cognet,
  title={CogNet: A Large-Scale Cognate Database},
  author={Batsuren, Khuyagbaatar and Bella, Gabor and Giunchiglia, Fausto},
  booktitle={Proceedings of the 57th Annual Meeting of the Association for Computational Linguistics},
  pages={3136--3145},
  year={2019}
}
```

Paper: [CogNet: A Large-Scale Cognate Database](https://aclanthology.org/P19-1302/) (ACL 2019, Florence, Italy)

## 🛠 Tech Stack

- Go with Fiber framework
- Redis for fast lookups
- Docker for easy setup

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.21 or higher

### Quick Start

1. Clone the repo
```bash
git clone https://github.com/yourusername/cognate-world-inquiry-service
cd cognate-world-inquiry-service
```

2. Start Redis
```bash
docker-compose up -d
```

3. Run the service
```bash
go run cmd/cognet-world-inquiry-service/main.go
```

## 📝 API Endpoints

### Import Data
```bash
# Import TSV file
POST /api/v1/import/tsv
```

### Search
```bash
# Get word suggestions
GET /api/v1/search/suggestions?prefix=bal

# Get cognates by concept ID
GET /api/v1/search/concept/{id}
```

## 📋 Example Responses

### Word Suggestions
```json
{
    "data": [
        {
            "word": "balık",
            "language": "tur",
            "concept_id": "n00001234"
        }
    ]
}
```

### Cognates by Concept ID
```json
{
    "data": [
        {
            "concept_id": "n00001234",
            "lang1": "tur",
            "word1": "balık",
            "lang2": "eng",
            "word2": "fish"
        }
    ]
}
```

## 🤝 Contributing

Feel free to open issues and submit PRs.

---
Made with 🎉 and Go
