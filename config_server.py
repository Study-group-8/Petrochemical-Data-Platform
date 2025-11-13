import os
import hashlib
from flask import Flask, jsonify
from flask_cors import CORS
from dotenv import load_dotenv

# Загрузка переменных окружения
load_dotenv()

app = Flask(__name__)
CORS(app)  # Разрешаем CORS для фронтенда

# Секретный ключ для хеширования
SECRET_KEY = os.getenv('SECRET_KEY', 'default-secret-key')
ADMIN_PASSWORD = os.getenv('ADMIN_EXPORT_PASSWORD', 'Petrochem2025!')


def hash_password(password: str) -> str:
    """Хеширует пароль с использованием SHA-256"""
    return hashlib.sha256(f"{password}{SECRET_KEY}".encode()).hexdigest()


@app.route('/api/config/password-hash', methods=['GET'])
def get_password_hash():
    """
    Возвращает хеш пароля для проверки на фронтенде
    Фронтенд будет хешировать введённый пароль и сравнивать с этим хешем
    """
    password_hash = hash_password(ADMIN_PASSWORD)
    return jsonify({
        'hash': password_hash,
        'algorithm': 'sha256'
    })


@app.route('/api/config/validate-password', methods=['POST'])
def validate_password():
    """
    Альтернативный метод - валидация пароля на сервере
    Более безопасный вариант
    """
    from flask import request
    
    data = request.get_json()
    input_password = data.get('password', '')
    
    is_valid = input_password == ADMIN_PASSWORD
    
    return jsonify({
        'valid': is_valid
    })


@app.route('/health', methods=['GET'])
def health():
    """Проверка здоровья сервиса"""
    return jsonify({
        'status': 'healthy',
        'service': 'config-server'
    })


if __name__ == '__main__':
    port = int(os.getenv('CONFIG_SERVER_PORT', 5000))
    app.run(host='0.0.0.0', port=port, debug=False)
