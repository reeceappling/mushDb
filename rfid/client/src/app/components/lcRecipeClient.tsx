'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import LiquidsArea, {
    IsValidLiquid,
    Liquid,
    LiquidEntriesGroup,
    LiquidEntriesGroupForNew
} from "@/app/components/formSubcomponents/liquids";
import NutrientsArea, {
    IsValidNutrient,
    Nutrient,
    NutrientEntriesGroup,
    NutrientsEntriesGroupForNew
} from "@/app/components/formSubcomponents/nutrients";
import SugarsArea, {
    IsValidSugar,
    Sugar,
    SugarEntriesGroup,
    SugarEntriesGroupForNew
} from "@/app/components/formSubcomponents/sugars";
import {
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
    RequiredArrayOfType
} from "@/app/components/common";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import AdditivesArea, {
    Additive,
    AdditiveEntriesGroup,
    AdditiveEntriesGroupForNew,
    IsValidAdditive
} from "@/app/components/formSubcomponents/additives";
import {ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {dataFor, ExistingDualSelector, InlineEntry} from "@/app/components/agarRecipeClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {NewLcForm} from "@/app/components/lcClient";
import {LcData} from "@/app/components/lcServer";
import {DepthContext, DepthProvider} from "./formSubcomponents/depthContext/depth";
import {
    FlexedArea,
    FlexedSinglesGroup,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea,
    NumberToDateStr
} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";

export function AssertLcRecipe(input: any): asserts input is LcRecipeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex required array keys
    let complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['liquids', IsValidLiquid],
    ])
    for (let [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            throw new Error('LcRecipe assertion failure: optional array key ' + key + ' was not valid. {' + JSON.stringify(input[key]) + '}{' + JSON.stringify(input) + '}');
        }
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['nutrients', IsValidNutrient],
        ['sugars', IsValidSugar],
        ['additives', IsValidAdditive],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('LcRecipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function LcRecipeDisplay(
    {
        id, readonly, data, headerLevel, cookies
    }: DisplayInput) {
    try {
        AssertLcRecipe(data)
        const [initial, setInitial] = useState(data)

        const [recName, setRecName] = useState(initial.name)
        const [isStandard, setIsStandard] = useState(initial.standard)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: LcRecipeData) => {
            setInitial(updated)
            setRecName(updated.name)
            setIsStandard(updated.standard)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const lcRecipeSubmit = () => {
            fetch(BaseExternalUrl + "/db/update/lcRecipe/" + initial._id, {
                method: "POST",
                headers: {
                    credentials: 'include',
                    'Content-type': "application/json"
                },
                body: JSON.stringify({
                    name: recName,
                    standard: isStandard,
                    notes: notes,
                    acl: MarshalAcl(acl),
                })
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertLcRecipe(entry)
                    updateInitial(entry)
                })
                .catch((error) => {
                    setErr(JSON.stringify(error))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "New LC from LcRecipe", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewLcForm lcRecipeIn={initial} handlers={{
                        onCreate: (newItem: LcData) => {
                            return onCreate([{
                                typeText: "Liquid Culture Jar", // TODO: validate ok
                                node: <CreatedLinkFor linkId={newItem._id} typ={"lc"}/>
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            }
            // TODO: any others?
        ]
        return (
            <DisplayFormWrapper entryType={"lcRecipe"}>
                <ErrorDisplay err={err}/>{/* TODO: OK?*/}
                <TestAndValidate todos={["Put name at top????"]}>
                    <ID id={data._id} txt={"Liquid Culture Recipe"} entryType={"lcRecipe"}/>
                </TestAndValidate>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <TestAndValidate todos={["allow name changes?"]}>
                            <NameArea currentName={data.name} readonly={readonly} headerLevel={headerLevel}
                                      setName={setRecName}/>
                        </TestAndValidate>

                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={readonly}
                                      headerLevel={headerLevel}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>


                <LiquidsArea initialValues={dataFor(initial.liquids || [])} headerLevel={headerLevel}
                             readonly={true}/>{/* TODO: viewOnlyArea*/}
                <NutrientsArea initialValues={dataFor(initial.nutrients || [])} headerLevel={headerLevel}
                               readonly={true}/>{/* TODO: viewOnlyArea*/}
                <SugarsArea initialValues={dataFor(initial.sugars || [])} headerLevel={headerLevel}
                            readonly={true}/>{/* TODO: viewOnlyArea*/}
                <AdditivesArea initialValues={dataFor(initial.additives || [])} headerLevel={headerLevel}
                               readonly={true}/>{/* TODO: viewOnlyArea*/}
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>

                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    lcRecipeSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Liquid Culture Recipe data format incorrect: " + err}</div>
    }
}

export function NewLcRecipeForm({handlers}: { handlers: NewEntryInput<LcRecipeData> }) {
    const [name, setName] = useState("")
    const [isStandard, setIsStandard] = useState(false)
    const [liquids, setLiquids] = useState<Liquid[]>([])
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    // TODO: handle isTopLevel
    const loadTemplate = (template: LcRecipeData) => {
        setLiquids(template.liquids)
        setNutrients(template.nutrients || [])
        setSugars(template.sugars || [])
        setAdditives(template.additives || [])
    }
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (name === "") {
            setErr("invalid name")
            return
        }
        fetch(BaseExternalUrl + "/db/create/lcRecipe", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
            body: JSON.stringify({
                name: name,
                standard: isStandard,
                liquids: liquids,
                nutrients: nutrients,
                sugars: sugars,
                additives: additives,
                notes: notes
            })
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertLcRecipe(newEntry)
                handlers.onCreate && handlers.onCreate(newEntry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            return <LcRecipeSelector doSelect={(rec) => { // TODO: endpoint for getStandard
                if (rec === undefined) {
                    return
                }
                loadTemplate(rec)
                setTemplateSelectorOpen(false)
            }} allowCreate={false}/>
        } else {
            return <button className={"basicButton"} onClick={() => {
                setTemplateSelectorOpen(true)
            }}>{"Select a template recipe (optional)"}</button>
        }
    }
    return (
        <NewEntryFormWrapper entryType={"lcRecipe"}>
            <ErrorDisplay err={err}/>
            <TestAndValidate todos={["TEST THIS"]}>
                {templateRecipeSelector()}
            </TestAndValidate>
            <NameArea classNames={"inlineChildren"} currentName={name} setName={setName} headerTxt={"Recipe name: "}
                      readonly={false}/>
            <StandardArea isStandard={isStandard} setStandard={setIsStandard} headerTxt={"Standard recipe? "}
                          readonly={false}/>
            <div>{"Liquids"}</div>
            <LiquidEntriesGroupForNew currentEntries={liquids} updateParent={setLiquids}/>
            <div>{"Nutrients"}</div>
            <NutrientsEntriesGroupForNew currentEntries={nutrients} updateParent={setNutrients}/>
            <div>{"Sugars"}</div>
            <SugarEntriesGroupForNew currentEntries={sugars} updateParent={setSugars}/>
            <div>{"Additives"}</div>
            <AdditiveEntriesGroupForNew currentEntries={additives} updateParent={setAdditives}/>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <button className={"greenButton buttonFullWidth"} onClick={createEntry}>{"Create"}</button>
        </NewEntryFormWrapper>
    )
}

function depthAndEntryClasses(depth: number, entryType?: string) {
    return " depth" + depth + (entryType ? " " + entryType : "")
}

// TODO: MOVE!
export function NewEntryFormWrapper(props: React.PropsWithChildren<{ entryType: string, className?: string }>) { // TODO: USE THIS EVERYWHERE!
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div
            className={"subForm newEntryForm" + depthAndEntryClasses(depth, props.entryType) + (props.className ? " " + props.className : "")}>{/* TODO: likely not working as expected. +1?*/}
            {props.children}
        </div>
    </DepthProvider>
}

// TODO: MOVE!
export function ImportEntryFormWrapper(props: React.PropsWithChildren<{ entryType: string }>) { // TODO: USE THIS EVERYWHERE!
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div
            className={"subForm importEntryForm" + depthAndEntryClasses(depth, props.entryType)}>{/* TODO: likely not working as expected. +1?*/}
            {props.children}
        </div>
    </DepthProvider>
}


// TODO: MOVE!
export function DisplayFormWrapper(props: React.PropsWithChildren<{ entryType: string, id?: string }>) { // TODO: USE THIS EVERYWHERE!
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div id={props.id}
             className={"subForm displayForm" + depthAndEntryClasses(depth, props.entryType)}>{/* TODO: likely not working as expected. +1?*/}
            {props.children}
        </div>
    </DepthProvider>
}

// TODO: MOVE!
export function Subform(props: React.PropsWithChildren<{}>) { // TODO: USE THIS EVERYWHERE!
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div className={"subForm depth" + depth}>{/* TODO: likely not working as expected. +1?*/}
            {props.children}
        </div>
    </DepthProvider>
}

export function LcRecipeInline({
                                   data,
                                   expandByDefault,
                                   onClick,
                                   showMainPageButton,
                                   idIsLink
                               }: InlineProps<LcRecipeData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Liquid Culture Recipe"} entryType={"lcRecipe"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            <NameArea currentName={data.name} headerTxt={"Recipe Name: "} readonly={true}/>
            <StandardArea isStandard={data.standard} headerTxt={"Standard? "} readonly={true}/>
            <LiquidEntriesGroup preexisting={true} readonly={true} initialEntries={data.liquids.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: Liquids (with more on expand)*/}
            <NutrientEntriesGroup preexisting={true} readonly={true} initialEntries={data.nutrients?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: NUTRIENTS (with more on expand)*/}
            <SugarEntriesGroup preexisting={true} readonly={true} initialEntries={data.sugars?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: SUGARS (with more on expand) */}
            <AdditiveEntriesGroup preexisting={true} readonly={true} initialEntries={data.additives?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: ADDITIVES (with more on expand) */}
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <NotesAreaInline notes={data.notes} header={"Notes: "} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                                                     expanded={expanded}/>
    </InlineEntry>
}

export function LcRecipeArea({lcRecipeId, headerLevel, offset}: {
    lcRecipeId?: string,
    headerLevel?: number,
    offset?: number
}) {
    // TODO: does this need incremented depth?
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (lcRecipeId !== undefined) {
        const b58id = lcRecipeId
        linkArea = <EntryLink props={{displayedId: b58id, linkId: b58id, entryType: "lcRecipe"}}>
            <div>{b58id}</div>
        </EntryLink>
    }
    return <div className={"lcRecipeArea"}>
        <div>{"Liquid Culture Recipe ID: "}</div>
        {linkArea}
    </div>
}

// export function LcRecipeSelector( // TODO: COLLAPSE! // TODO: likely overhaul to use the other List thingy!
//     {
//         doSelect, allowCreation, headerLevel, creatorInPage
//     }: SelectorProps<LcRecipeData>) {
//     // TODO: does this need incremented depth?
//     const [loaded, setLoaded] = useState(false)
//     const [open, setOpen] = useState(false)
//     const [selected, setSelected] = useState<LcRecipeData | undefined>()
//     const [standardList, setStandardList] = useState<LcRecipeData[] | undefined>()
//     const [recentList, setRecentList] = useState<LcRecipeData[] | undefined>()
//     const [err, setErr] = useState<string | undefined>()
//     //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
//     // TODO: RECIPE CREATOR SECTION!
//     useEffect(() => { // TODO: ENSURE WORKS
//         setSelected(undefined)
//         fetch(BaseExternalUrl + "/db/list/lcRecipes", {
//             method: "GET",
//             headers: {
//                 credentials: 'include',
//                 // 'Cookie': cookies, // TODO: do we even need this?
//                 // TODO: THIS!
//             },
//         })
//             .then(HandleJsonResponse)
//             .then((data) => { // TODO: DATA AS TWO LISTS!
//                 setStandardList(data.standard) // TODO: ASSERT????
//                 setRecentList(data.recent) // TODO: ASSERT????
//                 setLoaded(true)
//                 setErr(undefined)
//             })
//             .catch((error) => {
//                 setErr(JSON.stringify(error))
//             });
//     }, []);
//     const errArea = <ErrorDisplay err={err} headerLevel={headerLevel}/>
//     if (!loaded) {
//         return <div>{errArea}{"Loading LC Recipe Selector..."}</div>
//     }
//     if (!open) {
//         return <div>
//             {errArea}
//             {selected && <div>{"Recipe: " + selected._id}</div>}
//             <div>
//                 <button className={"basicButton"} onClick={() => {
//                     setOpen(true)
//                 }}>{selected ? "Select a different LC Recipe" : "Select an LC Recipe"}</button>
//             </div>
//         </div>
//     }
//     return <div>
//         {errArea}
//         <button className={"basicButton"} onClick={() => {
//             setOpen(false)
//         }}>{"Close Selector"}</button>
//         <DepthProvider>
//             <div>
//                 <div>{"Standard Recipes"}</div>
//                 {(standardList || []).map((recipe, i) => {
//                     return <div key={i} className={(selected && recipe._id === selected._id) ? "selectedItem" : ""}>
//                         <LcRecipeInline data={recipe} headerLevel={headerLevel} onClick={() => {
//                             doSelect(recipe)
//                             setSelected(recipe)
//                             setOpen(false)
//                         }}/>
//                     </div>
//                 })}
//             </div>
//             <div>
//                 <div>{"Recent Recipes"}</div>
//                 {(recentList || []).map((recipe, i) => {
//                     return <div className={(selected && recipe._id === selected._id) ? "selectedItem" : ""}>
//                         <LcRecipeInline data={recipe} headerLevel={headerLevel} onClick={() => {
//                             doSelect(recipe)
//                             setSelected(recipe)
//                             setOpen(false)
//                         }}/>
//                     </div>
//                 })}
//             </div>
//         </DepthProvider>
//         <button className={"basicButton"} onClick={() => {
//             setOpen(false)
//         }}>{"Close Selector"}</button>
//     </div>
// }

// export function LcRecipeListDisplay({recent, standard, onClick}: TwoListProps<LcRecipeData>) {
//     const recentArea = ()=>{
//         if(recent.length===0){
//             return null
//         }
//         return <div>
//             {standard.length>0 && <div>{"Recent Recipes:"}</div>}
//             {recent.map((b,i)=>{
//                 return <LcRecipeInline data={b} onClick={()=>{onClick(b)}} key={b._id}/>
//             })}
//         </div>
//     }
//     const standardArea = ()=>{
//         if(standard.length===0){
//             return null
//         }
//         return <div>
//             {recent.length>0 && <div>{"Standard Recipes:"}</div>}
//             {recent.map((b,i)=>{
//                 return <LcRecipeInline data={b} onClick={()=>{onClick(b)}} key={b._id}/>
//             })}
//         </div>
//     }
//     return <div>
//         {recentArea()}
//         {standardArea()}
//     </div>
// }

export function LcRecipeListPageTable({data, onClick, withLink}: ListPageItems<LcRecipeData>) {
    let cols: ListTableColumn<LcRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("Liquids", (v) => {
            return <div>
                {v.liquids.map((l, i) => {
                    return <div key={l.name + i}>{l.name}</div>
                })}
            </div>
        }),
        NewColumn("Nutrients", (v) => {
            return <div>
                {v.nutrients && v.nutrients.map((v, i) => {
                    return <div key={v.nutrient + i}>{v.nutrient}</div> // TODO: any more??
                })}
            </div>
        }),
        NewColumn("Sugars", (v) => {
            return <div>
                {v.sugars && v.sugars.map((v, i) => {
                    return <div key={v.type + i}>{v.type}</div> // TODO: any more??
                })}
            </div>
        }),
        NewColumn("Additives", (v) => {
            return <div>
                {v.additives && v.additives.map((v, i) => {
                    return <div key={v.additive + i}>{v.additive}</div> // TODO: any more??
                })}
            </div>
        }),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })
        // TODO: bonus area for notes???
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: LcRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "lcRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function LcRecipeSelectorTable({data, onClick}: ListPageItems<LcRecipeData>) {
    let cols: ListTableColumn<LcRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: LcRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "lcRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
        // TODO: bonus area for notes???
    ]
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function LcRecipeSelector( // TODO: USE ELSEWHERE
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: LcRecipeData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: LcRecipeData[]): JSX.Element => {
        return <LcRecipeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingDualSelector entryType={"lcRecipe"} entryTypes={"lcRecipes"} doSelect={doSelect}
                                 asserter={AssertLcRecipe}
                                 table={table}>
        {allowCreate && <NewLcRecipeForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingDualSelector>
}