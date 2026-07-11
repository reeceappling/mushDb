'use client'

import {useState} from "react";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {NoSsr} from "@mui/material";
import {InputNumber2, InputNumber4} from "@/app/components/formSubcomponents/numericInput";

interface DateProps {
    pre?: string,
    when?: number,
    readonly?: boolean,
    updateParent?: (ms: number) => void
}

export default function DateArea(
    {pre, when, readonly, updateParent}: DateProps) {
    const whenDate = new Date(when || Date.now())
    const [day, setDay] = useState<{ n: number, err: string | undefined }>({n: whenDate.getDate(), err: undefined})
    const [month, setMonth] = useState<{ n: number, err: string | undefined }>({n: whenDate.getMonth(), err: undefined}) // TODO: currently Jan==0?
    const [year, setYear] = useState<{ n: number, err: string | undefined }>({
        n: whenDate.getFullYear(),
        err: undefined
    })
    const updateDay = (s: string) => {
        const n = NumbersOnlyFromText(s)
        let err: string | undefined = undefined
        if (n < 1 || n > 31) {
            err = "Day must be [1,31]"
        } else {
            if (!dateValid(n, month.n, year.n)) {
                err = "invalid date"
            }
        }
        setDay({n: n, err: err})
        if (updateParent != undefined) {
            updateParent(new Date(year.n, month.n, n).getTime())
        }
    }
    const updateMonth = (s: string) => {
        const nInit = NumbersOnlyFromText(s)

        let err: string | undefined = undefined
        const n = nInit - 1
        if (nInit < 1 || nInit > 12) {
            err = "Month must be [1,12]"
        } else {
            if (!dateValid(day.n, n, year.n)) {
                err = "invalid date"
            }
        }

        setMonth({n: n, err: err})
        if (updateParent != undefined) {
            updateParent(new Date(year.n, n, day.n).getTime())
        }
    }
    const updateYear = (s: string) => {
        const n = NumbersOnlyFromText(s)
        let err: string | undefined = undefined
        if (n < 2024 || n > (new Date()).getFullYear()) {
            err = "invalid year, must be later than 2023 but not later than this year"
        }
        setYear({n: n, err: err})
        if (updateParent != undefined) {
                        updateParent(new Date(n, month.n, day.n).getTime())
        }
    }
    const setDateToNow = () => {
        const now = Date.now()
        const nowDate = new Date(now)
        setDay({n: nowDate.getDate(), err: undefined})
        setMonth({n: nowDate.getMonth(), err: undefined})
        setYear({n: nowDate.getFullYear(), err: undefined})
        if (updateParent != undefined) {
            updateParent(now)
        }
    }
    if (readonly) {
        if (when === undefined) { // WHEN MUST EXIST
            return <div className={"dateHolder"}>{(pre || "") + "none"}</div>
        }
        return <NoSsr>
            <div className={"dateHolder"}>
                {(pre !== "" && pre !== undefined) && <div className={"dateHeader"}>{pre}</div>}
                <div className={"dateValue"}>{NumberToDate(whenDate)}</div>
            </div>
        </NoSsr>
    }
    const currentErr = () => {
        if (day.err !== undefined) {
            return "Day Error: " + day.err
        }
        if (month.err !== undefined) {
            return "Month Error: " + month.err
        }
        if (year.err !== undefined) {
            return "Year Error: " + year.err
        }
        return undefined
    }
    // const dmyStr = (d:number,m:number,y:number)=>{
    //     let out = ""
    //     if (m<10){
    //         out=out+"0"
    //     }
    //     out=out+m+"/"
    //     if (d<10){
    //         out+="0"
    //     }
    //     out =out+d+"/"
    //     out=out+y
    //     return out
    // }
    return (
        <NoSsr>
            <div className={"dateHolder"}>
                {(pre !== "" && pre !== undefined) && <div>{pre}</div> /* TODO: P OK? */}
                <div className={'dateEditable'}>
                    <div className={"inlineChildren"}>
                        <div className={"inlineChildren inputNumbersInline"}>{/*TODO: MAYBE DELETE THIS WRAPPER DIV*/}
                            <InputNumber2 step={1} min={1} max={12} readonly={false} mode={"integer"}
                                          errorMessage={month.err} value={"" + (month.n + 1)} onChange={(e) => {
                                updateMonth(e || "")
                            }}/>
                            <InputNumber2 step={1} min={1} max={31} readonly={false} mode={"integer"}
                                          errorMessage={day.err} value={"" + day.n} onChange={(e) => {
                                updateDay(e || "")
                            }}/>
                            <InputNumber4 step={1} min={2000} max={10000} readonly={false} mode={"integer"}
                                          errorMessage={year.err} value={"" + year.n} onChange={(e) => {
                                updateYear(e || "")
                            }}/>
                        </div>
                        {/*<input className={"monthInput"} type="text" value={month.n+1} onChange={(e)=>{updateMonth(e.currentTarget.value)}}/>*/}
                        {/*<input className={"dayInput"} type="text" value={day.n} onChange={(e)=>{updateDay(e.currentTarget.value)}}/>*/}
                        {/*<input className={"yearInput"} type="text" value={year.n} onChange={(e)=>{updateYear(e.currentTarget.value)}}/>*/}
                        {!readonly &&
                            <button className={"basicButtonSmall"} onClick={setDateToNow}>{"Set date to now"}</button>}
                    </div>
                    <ErrorDisplay err={currentErr()}/>
                </div>


            </div>
        </NoSsr>
    )
}

export function NumberToDate(date: Date): string {
    return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
    }).format(date); // "01/12/2025"
}

export function NumbersOnlyFromText(s: string) {
    return Number(NumberStringOnlyFromText(s))
}

export function NumberStringOnlyFromText(s: string) {
    return s.replace(/\D/g,'')
}

function dateValid(dy: number, mo: number, yr: number) {
    const dt = new Date(yr, mo, dy)
    return dt.getFullYear() === yr &&
        dt.getMonth() === mo &&
        dt.getDate() === dy
}