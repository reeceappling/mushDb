'use client'

import React, {JSX, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedTriColFunction,
    AllEntries,
    Data,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
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
    HandleTxtResponse,
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
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {Grain, GrainsSelector, IsValidGrain} from "@/app/components/formSubcomponents/grains";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {ExistingDualSelector, InlineEntry} from "@/app/components/agarRecipeClient";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {NewGrainBatchForm} from "@/app/components/grainBatchClient";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
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


export function AssertJarRecipe(input: any): asserts input is JarRecipeData {
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
            throw new Error('Grain Jar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
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
        ['grains', IsValidGrain], // TODO: ensure length > 0
    ])
    for (let [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            throw new Error('Grain Jar Recipe assertion failure: required array key ' + key + ' was not valid');
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
            throw new Error('Grain Jar Recipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

function dataFormatFor<T>(init: T[] | undefined): Data<T>[] {
    return (init || []).map((v) => {
        return {data: v, disabled: false}
    })
}

export default function JarRecipeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertJarRecipe(data)
        const [initial, setInitial] = useState(data)
        // grain non-changeable (base grain)
        // name non-changeable
        const [isStandard, setIsStandard] = useState(initial.standard)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: JarRecipeData) => {
            setInitial(updated)
            setIsStandard(updated.standard)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const submit = () => {
            fetch(BaseExternalUrl + "/db/update/jarRecipe/" + data._id, { // TODO: ensure all using this dont start with /
                method: "POST",
                headers: {
                    credentials: 'include',
                    'Content-type': "application/json"
                },
                body: JSON.stringify({
                    standard: isStandard,
                    notes: notes,
                    acl: MarshalAcl(acl),
                })
            })
                .then(HandleTxtResponse)
                .then((newEntry) => {
                    AssertJarRecipe(newEntry)
                    updateInitial(newEntry)
                })
                .catch((error) => {
                    setErr(JSON.stringify(error))
                });
        }
        const jarGrainsArea = () => {
            return <div>
                {"Grains: "}
                {data.grains.map((g, i) => {
                    return <div key={g.grain}>{g.percentage + "% " + g.grain}</div>
                })}
            </div>
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "New Batch From Recipe", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewGrainBatchForm recipe={initial} handlers={{
                        onCreate: (newItem: GrainBatchData) => {
                            return onCreate([{
                                typeText: "Grain Batch",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"grainBatch"}/>
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            }
            // TODO: any others?
        ]
        return <DisplayFormWrapper entryType={"jarRecipe"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <TestAndValidate todos={["Put name at top????"]}>
                <ID id={data._id} txt={"Grain Jar Recipe"} entryType={"jarRecipe"}/>
            </TestAndValidate>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
            <FlexedArea>
                <FlexedSinglesGroup>
                    <TestAndValidate todos={["allow to be changeable?", "sometimes deletes name when doing updates"]}>
                        <NameArea currentName={initial.name} readonly={readonly} headerLevel={headerLevel}/>
                    </TestAndValidate>

                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    <StandardArea isStandard={isStandard} readonly={readonly} setStandard={setIsStandard}
                                  headerLevel={headerLevel}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            {jarGrainsArea()}

            <NutrientsArea initialValues={dataFormatFor(initial.nutrients)} headerLevel={headerLevel}
                           readonly={true}/>{/* TODO: make a viewOnlyArea?*/}
            <SugarsArea initialValues={dataFormatFor(initial.sugars)} headerLevel={headerLevel}
                        readonly={true}/>{/* TODO: make a viewOnlyArea?*/}
            <AdditivesArea initialValues={dataFormatFor(initial.additives)} headerLevel={headerLevel}
                           readonly={true}/>{/* TODO: make a viewOnlyArea?*/}
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}

        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Jar Recipe data format incorrect: " + err}</div>
    }
}

export function NewJarRecipeForm({handlers}: { handlers: NewEntryInput<JarRecipeData> }) {
    const [name, setName] = useState<string | undefined>()
    const [grains, setGrains] = useState<Grain[]>()
    const [isStandard, setIsStandard] = useState(false)
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    const loadTemplate = (template: JarRecipeData) => {
        setGrains(template.grains)
        setNutrients(template.nutrients || [])
        setSugars(template.sugars || [])
        setAdditives(template.additives || [])
    }
    const newJarRecipeSubmit = () => {
        if (!name) {
            setErr("Name must be set!")
            return
        }
        if (!grains || grains.length < 1) {
            setErr("Grains must be set!")
            return
        }
        let gs = grains || []
        let totalPct = 0
        for (let i = 0; i < gs.length; i++) {
            if (gs[i].percentage < 0 || gs[i].percentage > 100) {
                setErr("Invalid grain percentage")
            }
            totalPct += gs[i].percentage
        }
        if (totalPct != 100) {
            setErr("Grain percentages must equal 100")
            return
        }
        fetch(BaseExternalUrl + "/db/create/jarRecipe", {
            method: 'Post',
            body: JSON.stringify({
                name: name,
                grain: grains,
                standard: isStandard,
                nutrients: nutrients,
                sugars: sugars,
                additives: additives,
                notes: notes,
            }),
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertJarRecipe(newEntry)
                handlers.onCreate && handlers.onCreate(newEntry)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            return <JarRecipeSelector doSelect={(rec) => { // TODO: endpoint for getStandard?
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
    return <NewEntryFormWrapper entryType={"jarRecipe"}>
        <ErrorDisplay err={err}/>
        <TestAndValidate todos={["TEST THIS"]}>
            {templateRecipeSelector()}
        </TestAndValidate>
        <NameArea classNames={"inlineChildren"} readonly={false} setName={setName}/>

        <TestAndValidate todos={["ensure works like nutrientEntriesGroup"]}>
            <div>{"Grains"}</div>
            {/* TODO: grain batches???? */}
            {/* TODO: TITLE AREA AND MAKE THIS A COLUMN*/}
            <GrainsSelector current={grains || []} onChange={setGrains}/>
        </TestAndValidate>
        <StandardArea readonly={false} setStandard={setIsStandard}/>
        <div>{"Nutrients"}</div>
        <NutrientsEntriesGroupForNew currentEntries={nutrients} updateParent={setNutrients}/>
        <div>{"Sugars"}</div>
        <SugarEntriesGroupForNew currentEntries={sugars} updateParent={setSugars}/>
        <div>{"Additives"}</div>
        <AdditiveEntriesGroupForNew currentEntries={additives} updateParent={setAdditives}/>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton buttonFullWidth"} onClick={newJarRecipeSubmit}>{"Create Jar Recipe"}</button>
    </NewEntryFormWrapper>
}

export function JarRecipeInline({
                                    data,
                                    expandByDefault,
                                    onClick,
                                    showMainPageButton,
                                    idIsLink
                                }: InlineProps<JarRecipeData>) { // TODO: DO THIS ENTIRELY!
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Grain Jar Recipe"} entryType={"grainJar"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            <NameArea currentName={data.name} readonly={true}/>
            {/* TODO: ADD GRAINS */}
            <StandardArea isStandard={data.standard} readonly={true}/>
            <NutrientEntriesGroup preexisting={true} readonly={true} initialEntries={data.nutrients?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: NUTRIENTS (with more on expand)*/}
            <SugarEntriesGroup preexisting={true} readonly={true} initialEntries={data.sugars?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: Liquids (with more on expand)*/}
            <AdditiveEntriesGroup preexisting={true} readonly={true} initialEntries={data.additives?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: Additives (with more on expand) */}
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <NotesAreaInline notes={data.notes}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea>
        <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

export const JarRecipeArea = ({recipeId}: { recipeId?: string, headerLevel?: number }) => { // TODO: USE THIS!
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (recipeId !== undefined) {
        const b58id = recipeId
        linkArea = <EntryLink
            props={{displayedId: b58id, linkId: b58id, entryType: "jarRecipe"}}>{/* TODO: display name as well? */}
            <div>{b58id}</div>
            ]
        </EntryLink>
    }
    return <div>
        <div>{"Grain Recipe ID: "}</div>
        {linkArea}
    </div>
}

// export function JarRecipeSelector(
//     {doSelect, allowCreation, headerLevel, creatorInPage}: SelectorProps<JarRecipeData> // TODO: ALL PROPS
// ) {
//     // TODO; depth provider or no?
//     const [standard, setStandard] = useState<JarRecipeData[]>([])
//     const [recent, setRecent] = useState<JarRecipeData[]>([])
//     const [err, setErr] = useState<string | undefined>()
//     //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
//     useEffect(() => {
//         fetch(BaseExternalUrl + "/db/list/jarRecipes", {
//             method: 'GET',
//             headers: {
//                 credentials: 'include',
//                 // 'Cookie': cookies, // TODO: ok to leave out? do this elsewhere if it works
//                 'Content-type': "application/json"
//                 //Authorization: tokenFetch,
//             },
//         })
//             .then(HandleJsonResponse)
//             .then((resp) => {
//                 let out = resp as { standard: JarRecipeData[], recent: JarRecipeData[] } // TODO: ASSERT?
//                 setStandard(out.standard)
//                 setRecent(out.recent)
//             })
//             .catch((err) => {
//                 setErr(JSON.stringify(err))
//             });
//     }, []) // TODO: OK????? [] or nothing?
//     return <div>
//         <ErrorDisplay err={err} headerLevel={headerLevel}/>
//         <div>{/* Standard area*/}
//             <div>{"Standard Recipes"}</div>
//             {standard.map(item => {
//                 return <JarRecipeInline data={item} headerLevel={headerLevel} onClick={doSelect}/>
//             })}
//         </div>
//         <div>{/* Recent Area*/}
//             <div>{"Recent Recipes"}</div>
//             {recent.map(item => {
//                 return <JarRecipeInline data={item} headerLevel={headerLevel} onClick={doSelect}/>
//             })}
//         </div>
//         {/* TODO: CREATOR, IF ALLOWED, with increased depth */}
//     </div>
// }

// export function JarRecipeListDisplay({recent, standard, onClick}: TwoListProps<JarRecipeData>) {
//     const recentArea = () => {
//         if (recent.length === 0) {
//             return null
//         }
//         return <div>
//             {standard.length > 0 && <div>{"Recent Recipes:"}</div>}
//             {recent.map((b, i) => {
//                 return <JarRecipeInline data={b} onClick={() => {
//                     onClick(b)
//                 }} key={b._id}/>
//             })}
//         </div>
//     }
//     const standardArea = () => {
//         if (standard.length === 0) {
//             return null
//         }
//         return <div>
//             {recent.length > 0 && <div>{"Standard Recipes:"}</div>}
//             {recent.map((b, i) => {
//                 return <JarRecipeInline data={b} onClick={() => {
//                     onClick(b)
//                 }} key={b._id}/>
//             })}
//         </div>
//     }
//     return <div>
//         {recentArea()}
//         {standardArea()}
//     </div>
// }

export function JarRecipeListPageTable({data, onClick, withLink}: ListPageItems<JarRecipeData>) {
    let cols: ListTableColumn<JarRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("Grains", (v) => {
            return <div>
                {v.grains.map((g, i) => {
                    return <div key={g.grain + i}>{g.grain}</div>
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
        cols = [...cols, NewColumn("Link", (v: JarRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "jarRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function JarRecipeSelectorTable({data, onClick}: ListPageItems<JarRecipeData>) {
    let cols: ListTableColumn<JarRecipeData>[] = [
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("ID", (v) => v._id),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: JarRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "jarRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
        // TODO: bonus area for notes???
    ]
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function JarRecipeSelector( // TODO: USE ELSEWHERE
    {
        doSelect,
        allowCreate,
        creatorInPage,
    }: {
        doSelect: (val: JarRecipeData | undefined) => void,
        allowCreate?: boolean,
        creatorInPage?: boolean
    }) {
    const table = (items: JarRecipeData[]): JSX.Element => {
        return <JarRecipeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingDualSelector entryType={"jarRecipe"} entryTypes={"jarRecipes"} doSelect={doSelect}
                                 asserter={AssertJarRecipe}
                                 table={table}>
        {allowCreate && (creatorInPage ? <NewJarRecipeForm handlers={{onCreate: doSelect, isTopLevel: false}}/> :
            <div>{"LINK TO CREATOR HERE FIXME"}</div>)}
    </ExistingDualSelector>
}