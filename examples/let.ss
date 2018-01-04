(let ([a 6]
      [b 4])
  (eq? (+ a b) 10))

(let ([c 3])
  (let ([d 5])
    (eq? (+ c d) 8)))

(define e 2.7128)
(let ([f 3.1416])
  (+ e f))
