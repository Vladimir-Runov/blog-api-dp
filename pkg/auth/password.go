package auth

import (
	"crypto/rand"
	"errors"
	"math/big"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmptyPassword         = errors.New("password cannot be empty")
	ErrPasswordTooShort      = errors.New("password is too short")
	ErrPasswordLetterDigit   = errors.New("The password must contain at least one letter and one digit")
	ErrPasswordLetterCapital = errors.New("The password must contain at least one letter in lower and upper cases")
)

// HashPassword хеширует пароль используя bcrypt
func HashPassword(password string) (string, error) {
	// TODO: Реализовать хеширование пароля
	// Шаги:
	// 1. Проверить что пароль не пустой
	// 2. Использовать bcrypt для хеширования
	// 3. Выбрать подходящий cost factor (например, 10-12)
	// 4. Вернуть хешированный пароль как строку
	//
	// Подсказка: используйте golang.org/x/crypto/bcrypt
	// return "", errors.New("not implemented")

	if password == "" {
		return "", ErrEmptyPassword
	}
	if len(password) < 6 {
		return "", ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword проверяет соответствие пароля и его хеша
func CheckPassword(password, hash string) bool {
	// TODO: Реализовать проверку пароля
	// Шаги:
	// 1. Сравнить пароль с хешом используя bcrypt
	// 2. Вернуть true если пароль совпадает, false если нет
	// 3. При ошибке вернуть false
	//
	// Подсказка: bcrypt.CompareHashAndPassword

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true
	}
	return false
}

// ValidatePasswordStrength проверяет надежность пароля
func ValidatePasswordStrength(password string) error {
	// TODO: Реализовать проверку надежности пароля
	// Требования:
	// - Минимум 6 символов
	// - Опционально: содержит буквы и цифры
	// - Опционально: содержит заглавные и строчные буквы
	//
	// Вернуть соответствующую ошибку или nil
	// return errors.New("not implemented")

	if len(password) < 6 {
		return ErrPasswordTooShort
	}

	hasLetter := false
	hasDigit := false
	hasUpper := false
	hasLower := false

	// 2. Проверка на наличие букв и цифр
	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
			if unicode.IsUpper(char) {
				hasUpper = true
			} else {
				hasLower = true
			}
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return ErrPasswordLetterDigit
	}

	if !hasUpper || !hasLower {
		return ErrPasswordLetterCapital
	}

	// Если все проверки пройдены, возвращаем nil
	return nil

}

// GenerateRandomPassword генерирует случайный пароль (опциональное задание)
func GenerateRandomPassword(length int) (string, error) {
	if length < 6 {
		return "", errors.New("length must be at least 6")
	}

	lettersLc := "abcdefghijklmnopqrstuvwxyz"
	lettersUc := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"

	password := make([]byte, length)
	password[0], _ = randInt(lettersLc)
	password[1], _ = randInt(digits)
	password[2], _ = randInt(lettersUc)

	allChars := lettersUc + digits + lettersLc // Заполняем оставшуюся часть пароля случайными символами
	for i := 3; i < length; i++ {
		pchr, err1 := randInt(allChars)
		if err1 != nil {
			return "", err1
		}
		password[i] = pchr
	}
	shuffle(password)
	return string(password), nil
}

// randInt генерирует случайный символ из строки
func randInt(s string) (byte, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(s))))
	if err != nil {
		return 0, err
	}
	return s[nBig.Int64()], nil
}

// shuffle перемешивает массив байтов
func shuffle(slice []byte) {
	for i := len(slice) - 1; i > 0; i-- {
		j := randIntInRange(0, i+1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// randIntInRange генерирует случайное число в заданном диапазоне
func randIntInRange(min, max int) int {
	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return int(nBig.Int64()) + min
}
