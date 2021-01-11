Requirements: 
 - pass in functions at evaluation time
 - pass in arbitrary context to all functions
 - concurrent parsing 
 - eval.Vars() to optionally(?) include locality information. 
     - ex. a variable could understand which function, (and which argument) 
       it was a part of. 
     - each variable has a function to calculate "nearness"
 - marshalling of evaluatable object
   - parse should bring down the evaluatable expression to the smallest possible
     unit to evaluate. hence something like 3^5+1 should just be 244
 - work with: string, int64, float, bool 
   - straight-forward design to expand to support other datatypes
   - all types should implement all operators, but just 
     return errors if they are unapplicable
   - all types should define relationships to all other types for how/when 
     to convert between each other, when there is an error etc. 
 - monkey testing for panics 
 - ignoring/allowing for a preceding '=' sign
 - operators: 
   - string: `+` string concatination
   - math (bedmas): `(` `)` `+` `-` `*` `/` `^` `%` 
   - boolean: `true` `false` `&&` `||` `>` `<` `<=` `>=` `!=` `!`
   - Ternary conditional: `?` `:`
   - functions: `yourfnname(...)`

Parsing: 
 - first the expression is seperated out into an array of all elements. 
 - each element based on it's surroundings is either waiting or begins concurrently executing (in a certain direction)
 - upon execution in a first direction, signals are sent outward in both directions as to the element positions 
   which are executed, thier value, and new surrounding element positions. 

   r=ready, sends its value, and the oppoite directions element kind
            if the element is the last element, it will send an "evaluation-complete"
            signal leftward
   w=waiting, becomes a variable element once both sides information 
              has been sent in. the variable will have the leftward and rightward
              elements as has been provided by the signals sent in. 
   wl= waiting only on the left side
   wr= waiting only on the right side
   t=transmitter, once an element has sent its signal, it simply 
                  becomes a transmitter to the signals which pass in 
                  its direction. during implementation, 
                  rather than actually receiving and re-sending signals, 
                  priority as the recieving chanels should just be included based
                  on element position, and transmitters will simply be null. 
   e=evaluted result signal sent. value as well as opposite direction element
   c=completed signal

// TODO rework this example with information from below. 
```
example-1
=   A      +     B    /     C    *    4  [end]
wr  r->    w     r->  w   <-r    w  <-r  wlr
    t      wr    t    e->   t    wl   t
           wr         t        <-e
wr       <-e                     t
c   
```

how should priority be implmented however?
what if two elements as stuck on "waiting-rightward" and then a rightward 
signal gets sent? is it enough to send both the position and the element of the opposite 
direction with signals? this way when the signal gets sent, it gets sent to the specific chanel waiting for it?
maybe when the 'r' or 'e' sends its value, it also sends the opposite channel directly... hence the same channel object
can now be owned by the reciever of said signal (from 'r' or 'e'). beautiful. 

 
wrr= waiting-right, reflect signal, when the signal is recieved, 
     reflect its new signal rightward.
wrl= waiting-left, reflect signal 


```
example 2
=    someFn(  A    ,    B    ,    C    )    [end]
wr      wrr   r->  w    r->  w  <-r    wl    wlr
              t    wr   t  <-e    t    wl
        wrr      <-e
         e->                           wl 
         t                              e->  wlr
wr                                      t  <-c 
c
```


interesting things to note
 - the first variable after the `=` will only ever wait on the right hand side
 - for a waiting element, when it receives its final signal, it will send its new signal
   in the same direction as the final signal it received. SCRATCH THAT, not true, the direction
   is determined by the priotity of the surrounding elements (or virtual surrounding elements 
   as passed in by the input signals to the operator).  need to define this priority
   map, including `[end]` and `=` operators
 - elements can either be one of two types, operator or variable. An operator
   element is converted into a variable element once it has been evaluated. 
 - the `[end]` element will bounce signal directions 
 - the 'evaluated value' of each item in the comma series will stack on top of each 
   other meaning that the 'value' which gets sent into the function will be an 
   array of all the stacked comma items. this is also why the first `,` in example 2
   needs to be left in the `wr` state to wait for the rightward items to be sent 
   towards it to then "stack-on" its leftward item. 

-----------------------------------
Preprocessing:
 - add * symbols between any vars and bracket outsides
 - add * symbols between vars (would have been seperated by spaces)
 - remove a preceding `=` if it exists 
 - error if uneven number of `?` and `:` symbols
 - error if uneven number of brackets leftwise and rightwise
 - parse between modifier and operator versions of `-` 
 - error on double modifier
 - error if an [non-modifier] operator exists on the inside side of a bracket
   - brackets on brackets are allowed

Object Types: IMPORTANT REFERENCE : https://en.wikipedia.org/wiki/Order_of_operations
 - Variables 
   - send thier information in the priority direction
   - `A` 
   - always send to a [`^`] operator before the `-` modifier
   - always send to the `!` modifier before another logical operator
   - always send to the `-` modifier before another [`+` `-` `*` `/`] operator
 - Operators 
   - HAVE RECEIVERS which at init are sent to surrounding VARS, MODS, and BRACS
   - collect input from both sides then execute and become a variable
   - `,` `+` `-` `*` `/` `^` `%` `&&` `||` `>` `<` `<=` `>=` `!=`
   - `&&` operator has higher priority than `||` 
   - comparors (`>` `<` `<=` `>=` `!=`) have a higher priority than 
     logical operators (`&&` `||`)
   - Tertiary Operator 
     - `?` `:`
     - both have lower priority than all other operators (`?` then `:`) 
     - `?` operates like a regular operator, if left is true, var is right 
           if left is false var becomes special var "tertiary-false" type
     - `:` operates like a regular operator, if left is "tertiary-false" then
           becomes right, anything else var becomes left. 
           All other operators error if they hit the "tertiary-false" type
 - Modifiers 
   - HAVE RECEIVERS which at init are sent to surrounding VARS and BRACS
   - recieves variables from the right side then becomes a variable
   - `-` `!`
 - Brackets (Are this kind of just variables?) 
   - HAVE RECEIVERS which at init are sent INWARD to the surrounding VARS and BRACS
   - `someFunc(` `(` `)`
   - variables will always send away from a bracket unless there are brackets
     on each side, in which case they dissolve both brackets (calling the function
     in the process if there is one) then becomme the super-variable.
   - variables need to be able to "reach" into each bracket to gain the surrounding 
     destination channels from outside each bracket. 
   - have a "suck in" function to take on the brackets outer channels?

DO NOT use [start] and [end] operators. just have nil operators, if a variable has 2 nil 
operators, conclude the calculation. 

PRIORITY ORDERING
 - `^` 
 - `-` modifier 
 - `!` modifier 
 - `*` `/` 
 - `%` 
 - `+` `-` operator
 - `>` `<` `<=` `>=` `!=` `==`
 - `&&` `||`
 - `?` 
 - `:` 
 - `,`
 - `(` `)`
