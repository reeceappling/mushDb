// TODO: From: https://dev.to/morewings/lets-create-a-better-number-input-with-react-1j0m

import {ChangeEvent, FC, KeyboardEvent, useEffect, useState} from 'react';
import {useCallback, useId} from 'react';
import * as React from "react";

// List of available numeric modes
export enum Modes {
    natural = 'natural',
    integer = 'integer',
    floating = 'floating',
    scientific = 'scientific',
}

type Value = string;

export interface ReadOnly {
    readonly?: boolean
}
export interface Labelled {
    /** Attach a text label to the input */
    label?: string;
}

export interface NumericAreaProps extends ReadOnly, NumericInputProps {
}

export interface NumericInputProps extends Labelled, NumericInputOnlyProps {
}
export interface NumericInputProps2 extends Labelled, NumericInputOnlyProps {
}
export type NumericInputOnlyProps = {
    key?:string;
    /** Read only? **/
    readonly?: boolean
    /** Set controlled value */
    value?: Value;
    /** Provide a callback to capture changes */
    onChange?: (value?: Value) => void;
    /**
     * Define a number to increase or decrease input value
     * when user clicks arrow keys
     */
    step?: number;
    /** Set a maximum value available for arrow stepping */
    max?: number;
    /** Set a minimum value available for arrow stepping */
    min?: number;
    /** Select a mode of numeric input */
    mode?: keyof typeof Modes;
    /** Set at a placeholder text for the input */
    placeholder?: string;
    /** Provide an error hint for the user*/
    errorMessage?: string;
};
export interface TextInputProps extends Labelled, TextInputOnlyProps {
}
export type TextInputOnlyProps = {
    /** Read only? **/
    readonly?: boolean
    /** Set controlled value */
    value?: Value;
    /** Provide a callback to capture changes */
    onChange?: (value?: Value) => void;
    /** Set at a placeholder text for the input */
    placeholder?: string;
    /** Provide an error hint for the user*/
    errorMessage?: string;
    onBlur?: () => void;
};

const patternMapping = {
    [Modes.natural]: '(?:0|[1-9]\\d*)',
    [Modes.integer]: '[+\\-]?(?:0|[1-9]\\d*)',
    [Modes.floating]: '[+\\-]?(?:0|[1-9]\\d*)(?:\\.\\d*)?',
    [Modes.scientific]: '[+\\-]?(?:0|[1-9]\\d*)(?:\\.\\d+)?(?:[eE][+\\-]?\\d+)?',
};

export const NumericalArea: FC<NumericAreaProps> = (
    props) => {
    if (props.readonly) {
        return DisplayNumerical(props)
    }
    return InputNumerical(props)
}
export const NumericalAreaWithAbsolutes: FC<NumericAreaProps> = (
    props) => {
    if (props.readonly) {
        return DisplayNumerical2(props)
    }
    return InputNumerical2(props)
}

const numericFieldClasses = "numericField px-2 translate-y-5"
const textFieldClasses = "textField px-2 translate-y-5"
const numInputClassName = "peer block w-48 rounded-none border-2 border-gray-300 bg-input px-4 text-right text-sm font-normal tabular-nums text-gray-900 placeholder:text-gray-400 invalid:border-red-600 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"
const txtInputClassName = "peer block w-48 rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal tabular-nums text-gray-900 placeholder:text-gray-400 invalid:border-red-600 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"
const errClassName = "invisible text-xs text-red-600 peer-[:invalid]:visible"
const errClassName2 = "invisible absolute top-[100%] ml-2 w-46 text-xs text-red-600 peer-[:invalid]:visible"
const labelClassCommon = "cursor-pointer items-center text-sm font-medium text-gray-600"
const labelClassAbsolute = "absolute top-[-1.25rem] "+labelClassCommon
const labelClassInline = ""+labelClassCommon

export const DisplayNumerical: FC<NumericInputProps> = ({
                                                            value,
                                                            step = 1,
                                                            max = Infinity,
                                                            min = -Infinity,
                                                            onChange = () => {
                                                            },
                                                            mode = Modes.floating,
                                                            label = 'Numeric input with default label',
                                                            placeholder, // placeholder number
                                                            errorMessage = 'error!',
                                                        }) => {
    const id = useId();
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    const pattern = patternMapping[mode];
    return (
        <fieldset className={numericFieldClasses}>
            <label
                htmlFor={id}
                className={labelClassAbsolute}
            >{label}</label>
            <input // TODO: USE SOMETHING OTHER THAN INPUT!!!!
                inputMode="decimal"
                autoComplete="off"
                pattern={pattern}
                onChange={handleChange}
                value={value !== undefined ? value : ''}
                type="text"
                id={id}
                className={numInputClassName}
                placeholder={placeholder}
                aria-describedby={`${id}-helper-text`}
            />
            <div
                className={errClassName}
                id={`${id}-helper-text`} // TODO: ?????????????
            >{errorMessage}</div>
        </fieldset>
    );
};
export const DisplayNumerical2: FC<NumericInputProps> = ({
                                                             value,
                                                             step = 1,
                                                             max = Infinity,
                                                             min = -Infinity,
                                                             onChange = () => {
                                                             },
                                                             mode = Modes.floating,
                                                             label = 'Numeric input with default label',
                                                             placeholder, // placeholder number
                                                             errorMessage = 'error!',
    key
                                                         }) => {
    const id = useId();
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    const pattern = patternMapping[mode];
    return (<>
            <label key={key+"lab"}
                htmlFor={id}
                className={labelClassAbsolute}
            >{label}</label>
            <div key={key+"div1"} className={"inputAndErrWrapper"}>
                <input
                    inputMode="decimal"
                    autoComplete="off"
                    pattern={pattern}
                    onChange={handleChange}
                    value={value !== undefined ? value : ''}
                    type="text"
                    id={id}
                    className={numInputClassName}
                    placeholder={placeholder}
                    aria-describedby={`${id}-helper-text`}
                />
                <div
                    className={errClassName2}
                    id={`${id}-helper-text`} // TODO: ?????????????
                >{errorMessage}</div>
            </div>
        </>
    );
};
export const InputNumerical: FC<NumericInputProps> = (
    {
        readonly = false,
        value,
        step = 1,
        max = Infinity,
        min = -Infinity,
        onChange = () => {
        },
        mode = Modes.floating,
        label = 'Numeric input with default label',
        placeholder, // placeholder number
        errorMessage = 'error!',
    }) => {
    const id = useId();
    const handleKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            const inputValue = (event.target as HTMLInputElement).value;
            if (event.key === 'ArrowUp') {
                const nextValue = Number(inputValue || 0) + step;
                if (nextValue <= max) {
                    onChange(nextValue.toString());
                }
            }
            if (event.key === 'ArrowDown') {
                const nextValue = Number(inputValue || 0) - step;
                if (nextValue >= min) {
                    onChange(nextValue.toString());
                }
            }
        },
        [max, min, onChange, step]
    );
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    const pattern = patternMapping[mode];
    return (
        <fieldset className={numericFieldClasses}>
            <label
                htmlFor={id}
                className={labelClassAbsolute}
            >{label}</label>
            <input
                disabled={readonly}
                inputMode="decimal"
                autoComplete="off"
                pattern={pattern}
                onChange={handleChange}
                onKeyDown={handleKeyDown}
                value={value !== undefined ? value : ''}
                type="text"
                id={id}
                className={numInputClassName}
                placeholder={placeholder}
                aria-describedby={`${id}-helper-text`}
            />
            <div
                className={errClassName}
                id={`${id}-helper-text`}
            >{errorMessage}</div>
        </fieldset>
    );
};
export const InputNumerical2: FC<NumericInputProps> = (
    {
        readonly = false,
        value,
        step = 1,
        max = Infinity,
        min = -Infinity,
        onChange = () => {
        },
        mode = Modes.floating,
        label = 'Numeric input with default label',
        placeholder, // placeholder number
        errorMessage = 'error!',
        key
    }) => {
    const id = useId();
    const handleKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            const inputValue = (event.target as HTMLInputElement).value;
            if (event.key === 'ArrowUp') {
                const nextValue = Number(inputValue || 0) + step;
                if (nextValue <= max) {
                    onChange(nextValue.toString());
                }
            }
            if (event.key === 'ArrowDown') {
                const nextValue = Number(inputValue || 0) - step;
                if (nextValue >= min) {
                    onChange(nextValue.toString());
                }
            }
        },
        [max, min, onChange, step]
    );
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    const pattern = patternMapping[mode];
    return (
        <div key={key} className={"relative"}>
            <label
                htmlFor={id}
                className={labelClassAbsolute}
            >{label}</label>
            <input
                disabled={readonly}
                inputMode="decimal"
                autoComplete="off"
                pattern={pattern}
                onChange={handleChange}
                onKeyDown={handleKeyDown}
                value={value !== undefined ? value : ''}
                type="text"
                id={id}
                className={numInputClassName}
                placeholder={placeholder}
                aria-describedby={`${id}-helper-text`}
            />
            <div
                className={errClassName2}
                id={`${id}-helper-text`} // TODO: ?????????????
            >{errorMessage}</div>
        </div>
    )
        ;
};
export function InputDecimal({initial,label,min,max,updateParent}:{initial:number,label:string,min?:number,max?:number,updateParent:(n:number)=>void}){
    const [radiusDraft, setRadiusDraft] = useState(initial.toString())
    const [err, setErr] = useState<string | undefined>(undefined)
    useEffect(() => {
        setRadiusDraft(initial.toString())
    }, [initial])
    return <NumericalAreaWithAbsolutes label={label} mode="floating" min={min||0.0} max={max||1000.0} readonly={false}
                                       errorMessage={err} value={radiusDraft}
                                       onChange={(val?: string) => {
                                           const next = val ?? ""
                                           setRadiusDraft(next)
                                           // allow in-progress values like "1."
                                           if (next === "" || next.endsWith(".")) {
                                               setErr(undefined)
                                               return
                                           }
                                           try {
                                               const n = Number(val) // TODO: allow only numbers here
                                               if (!Number.isNaN(n)) {
                                                   val && updateParent(n)
                                                   setErr(undefined)
                                               } else {
                                                   setErr("NaN decimal input")
                                               }
                                           } catch (e) {
                                               setErr("failed to set decimal state: "+JSON.stringify(e))
                                           }
                                       }}/>
};

export const InputTextWithSmallTitle: FC<TextInputProps> = (
    {
        readonly = false,
        value,
        onChange = () => {
        },
        label = 'Text input with default label',
        placeholder, // placeholder text
        errorMessage = 'error!',
    }) => {
    const id = useId();
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    return <div className={"relative"}>
        <label
            htmlFor={id}
            className={labelClassAbsolute}
        >{label}</label>
        <InputText placeholder={placeholder} value={value} onChange={onChange} readonly={readonly} errorMessage={errorMessage}/>
        {/* TODO: revert if not working <InputText placeholder={placeholder} value={value} onChange={onChange} readonly={readonly} errorMessage={errorMessage}/>*/}
    </div>
};

export const InputTextInlineTitle: FC<TextInputProps> = (
    {
        readonly = false,
        value,
        onChange = () => {
        },
        label = 'Text input with default label',
        placeholder, // placeholder text
        errorMessage = 'error!',
    }) => {
    const id = useId();
    const handleChange = useCallback( // TODO: use?
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    return <div className={"inlineChildren"}>
        <label htmlFor={id} className={"text-m"}>{label}</label>
        <div className={"relative"}>
            <InputText placeholder={placeholder} value={value} onChange={onChange} readonly={readonly} errorMessage={errorMessage}/>
        </div>
    </div>
};

export const InputText: FC<TextInputOnlyProps> = (
    {
        readonly = false,
        value,
        onChange = () => {
        },
        placeholder, // placeholder text
        errorMessage = 'error!',
        onBlur,
    }) => {
    const id = useId();
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    return <>
            <input
                type="text"
                disabled={readonly}
                autoComplete="off"
                onChange={handleChange}
                value={value !== undefined ? value : ''}

                id={id}
                className={numInputClassName}
                placeholder={placeholder}
                aria-describedby={`${id}-helper-text`}
                onBlur={e=>{onBlur && onBlur()}} // TODO: ensure ok
            />
            <div
                className={errClassName2}
                id={`${id}-helper-text`} // TODO: ?????????????
            >{errorMessage}</div>
    </>
};

export const InputNumber: FC<NumericInputOnlyProps> = (
    {
        readonly = false,
        value,
        step = 1,
        max = Infinity,
        min = -Infinity,
        onChange = () => {
        },
        mode = Modes.floating,
        placeholder, // placeholder number
        errorMessage = 'error!',
    }) => {
    const id = useId();
    const handleKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            const inputValue = (event.target as HTMLInputElement).value;
            if (event.key === 'ArrowUp') {
                const nextValue = Number(inputValue || 0) + step;
                if (nextValue <= max) {
                    onChange(nextValue.toString());
                }
            }
            if (event.key === 'ArrowDown') {
                const nextValue = Number(inputValue || 0) - step;
                if (nextValue >= min) {
                    onChange(nextValue.toString());
                }
            }
        },
        [max, min, onChange, step]
    );
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    const pattern = patternMapping[mode];
    return <>
        <input
            disabled={readonly}
            inputMode="decimal"
            autoComplete="off"
            pattern={pattern}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            value={value !== undefined ? value : ''}
            type="text"
            id={id}
            className={numInputClassName}
            placeholder={placeholder}
            aria-describedby={`${id}-helper-text`}
        />
        <div
            className={errClassName2}
            id={`${id}-helper-text`} // TODO: ?????????????
        >{errorMessage}</div>
    </>
};

export const InputNumber2: FC<NumericInputOnlyProps> = (
    {
        readonly = false,
        value,
        step = 1,
        max = Infinity,
        min = -Infinity,
        onChange = () => {
        },
        mode = Modes.floating,
        placeholder, // placeholder number
        errorMessage = 'error!',
    }) => {
    const id = useId();
    const handleKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            const inputValue = (event.target as HTMLInputElement).value;
            if (event.key === 'ArrowUp') {
                const nextValue = Number(inputValue || 0) + step;
                if (nextValue <= max) {
                    onChange(nextValue.toString());
                }
            }
            if (event.key === 'ArrowDown') {
                const nextValue = Number(inputValue || 0) - step;
                if (nextValue >= min) {
                    onChange(nextValue.toString());
                }
            }
        },
        [max, min, onChange, step]
    );
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    const pattern = patternMapping[mode];
    return <>
        <input
            disabled={readonly}
            inputMode="decimal"
            autoComplete="off"
            pattern={pattern}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            value={value !== undefined ? value : ''}
            type="text"
            id={id}
            className={"peer block w-7 rounded-none border-2 border-gray-300 bg-input px-1 text-right text-sm font-normal tabular-nums text-gray-900 placeholder:text-gray-400 invalid:border-red-600 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
            placeholder={placeholder}
            aria-describedby={`${id}-helper-text`}
        />
        <div
            className={errClassName2}
            id={`${id}-helper-text`} // TODO: ?????????????
        >{errorMessage}</div>
    </>
};

export const InputNumber4: FC<NumericInputOnlyProps> = (
    {
        readonly = false,
        value,
        step = 1,
        max = Infinity,
        min = -Infinity,
        onChange = () => {
        },
        mode = Modes.floating,
        placeholder, // placeholder number
        errorMessage = 'error!',
    }) => {
    const id = useId();
    const handleKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            const inputValue = (event.target as HTMLInputElement).value;
            if (event.key === 'ArrowUp') {
                const nextValue = Number(inputValue || 0) + step;
                if (nextValue <= max) {
                    onChange(nextValue.toString());
                }
            }
            if (event.key === 'ArrowDown') {
                const nextValue = Number(inputValue || 0) - step;
                if (nextValue >= min) {
                    onChange(nextValue.toString());
                }
            }
        },
        [max, min, onChange, step]
    );
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            onChange(event.target.value);
        },
        [onChange]
    );
    const pattern = patternMapping[mode];
    return <>
        <input
            disabled={readonly}
            inputMode="decimal"
            autoComplete="off"
            pattern={pattern}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            value={value !== undefined ? value : ''}
            type="text"
            id={id}
            className={"peer block w-11 rounded-none border-2 border-gray-300 bg-input px-1 text-right text-sm font-normal tabular-nums text-gray-900 placeholder:text-gray-400 invalid:border-red-600 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
            placeholder={placeholder}
            aria-describedby={`${id}-helper-text`}
        />
        <div
            className={errClassName2}
            id={`${id}-helper-text`} // TODO: ?????????????
        >{errorMessage}</div>
    </>
};